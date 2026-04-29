package feed

import (
	"context"
	"sync"
	"time"

	pbFeed "ouroboros/proto/generated/feed"
)

type FanoutEngine struct {
	queue chan func()
	rate  <-chan time.Time
}

func NewFanoutEngine() *FanoutEngine {
	queue := make(chan func(), 1000)

	engine := &FanoutEngine{
		queue: queue,
		rate:  time.Tick(2 * time.Millisecond),
	}

	for i := 0; i < 16; i++ {
		go func() {
			for job := range queue {
				job()
			}
		}()
	}

	return engine
}

func (f *FanoutEngine) Run(ctx context.Context, store *Store, job *FanoutJob, item *pbFeed.FeedItem) error {
	total := (len(job.Followers) + job.BatchSize - 1) / job.BatchSize

	sem := make(chan struct{}, 8)
	errCh := make(chan error, total)

	var mu sync.Mutex

	for i := 0; i < total; i++ {
		if job.Completed[i] {
			continue
		}

		sem <- struct{}{}

		start := i * job.BatchSize
		end := start + job.BatchSize
		if end > len(job.Followers) {
			end = len(job.Followers)
		}

		batch := job.Followers[start:end]

		f.queue <- func() {
			defer func() { <-sem }()
			<-f.rate

			if err := store.FanoutBatch(ctx, batch, item); err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			job.Completed[i] = true
			_ = store.SaveJob(ctx, job)
			mu.Unlock()
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
