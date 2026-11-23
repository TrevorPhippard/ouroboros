package dataloader

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type recorder struct {
	mu      sync.Mutex
	batches [][]int
}

func (r *recorder) fetch(_ context.Context, keys []int) (map[int]string, error) {
	r.mu.Lock()
	cp := append([]int(nil), keys...)
	sort.Ints(cp)
	r.batches = append(r.batches, cp)
	r.mu.Unlock()

	out := make(map[int]string, len(keys))
	for _, k := range keys {
		if k >= 0 { // negative keys are "not found"
			out[k] = "v" + string(rune('a'+k%26))
		}
	}
	return out, nil
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.batches)
}

func TestConcurrentLoadsShareOneBatch(t *testing.T) {
	rec := &recorder{}
	l := New(context.Background(), Config{Wait: 20 * time.Millisecond}, rec.fetch)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got, err := l.Load(context.Background(), i)
			if err != nil || got == "" {
				t.Errorf("Load(%d) = %q, %v", i, got, err)
			}
		}(i)
	}
	wg.Wait()

	if n := rec.count(); n != 1 {
		t.Fatalf("expected 1 batch, got %d: %v", n, rec.batches)
	}
	if len(rec.batches[0]) != 10 {
		t.Fatalf("expected 10 keys in batch, got %v", rec.batches[0])
	}
}

func TestDuplicateKeysAreCollapsedWithoutCache(t *testing.T) {
	rec := &recorder{}
	l := New(context.Background(), Config{Wait: 10 * time.Millisecond, DisableCache: true}, rec.fetch)

	got, err := l.LoadMany(context.Background(), []int{1, 1, 2, 2, 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %v", got)
	}
	if rec.count() != 1 || len(rec.batches[0]) != 2 {
		t.Fatalf("expected one batch of 2 unique keys, got %v", rec.batches)
	}
}

func TestCacheServesRepeatLoads(t *testing.T) {
	rec := &recorder{}
	l := New(context.Background(), Config{Wait: time.Millisecond}, rec.fetch)

	for i := 0; i < 3; i++ {
		if _, err := l.Load(context.Background(), 7); err != nil {
			t.Fatal(err)
		}
	}
	if rec.count() != 1 {
		t.Fatalf("expected the cache to prevent refetching, got %d batches", rec.count())
	}

	l.Clear(7)
	if _, err := l.Load(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if rec.count() != 2 {
		t.Fatalf("expected Clear to force a refetch, got %d batches", rec.count())
	}
}

func TestDisableCacheRefetches(t *testing.T) {
	rec := &recorder{}
	l := New(context.Background(), Config{Wait: time.Millisecond, DisableCache: true}, rec.fetch)
	for i := 0; i < 3; i++ {
		if _, err := l.Load(context.Background(), 7); err != nil {
			t.Fatal(err)
		}
	}
	if rec.count() != 3 {
		t.Fatalf("expected 3 batches with cache disabled, got %d", rec.count())
	}
}

func TestMissingKeyIsZeroValueNotError(t *testing.T) {
	rec := &recorder{}
	l := New(context.Background(), Config{Wait: time.Millisecond}, rec.fetch)

	got, err := l.Load(context.Background(), -1)
	if err != nil || got != "" {
		t.Fatalf("expected zero value and nil error, got %q, %v", got, err)
	}
	many, err := l.LoadMany(context.Background(), []int{-1, 3})
	if err != nil || len(many) != 2 || many[-1] != "" || many[3] == "" {
		t.Fatalf("unexpected LoadMany result %v, %v", many, err)
	}
}

func TestPrimeSeedsCacheAndDoesNotOverwrite(t *testing.T) {
	rec := &recorder{}
	l := New(context.Background(), Config{Wait: time.Millisecond}, rec.fetch)

	l.Prime(5, "primed")
	l.Prime(5, "ignored")
	got, err := l.Load(context.Background(), 5)
	if err != nil || got != "primed" {
		t.Fatalf("got %q, %v", got, err)
	}
	if rec.count() != 0 {
		t.Fatalf("primed key must not hit the batch function")
	}
}

func TestErrorsAreSharedButNotCached(t *testing.T) {
	var calls atomic.Int32
	boom := errors.New("boom")
	l := New(context.Background(), Config{Wait: 10 * time.Millisecond}, func(_ context.Context, keys []int) (map[int]string, error) {
		if calls.Add(1) == 1 {
			return nil, boom
		}
		return map[int]string{keys[0]: "ok"}, nil
	})

	t1, t2 := l.LoadThunk(1), l.LoadThunk(2)
	if _, err := t1(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("expected boom, got %v", err)
	}
	if _, err := t2(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("expected boom for sibling key, got %v", err)
	}

	got, err := l.Load(context.Background(), 1)
	if err != nil || got != "ok" {
		t.Fatalf("expected retry to succeed, got %q, %v", got, err)
	}
}

func TestPanicBecomesError(t *testing.T) {
	l := New(context.Background(), Config{Wait: time.Millisecond}, func(context.Context, []int) (map[int]string, error) {
		panic("kaboom")
	})
	if _, err := l.Load(context.Background(), 1); err == nil {
		t.Fatal("expected an error from a panicking batch function")
	}
}

func TestMaxBatchDispatchesImmediately(t *testing.T) {
	rec := &recorder{}
	// A huge wait proves that dispatch was triggered by size, not by the timer.
	l := New(context.Background(), Config{Wait: time.Hour, MaxBatch: 3}, rec.fetch)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got, err := l.LoadMany(ctx, []int{1, 2, 3})
	if err != nil || len(got) != 3 {
		t.Fatalf("got %v, %v", got, err)
	}
	if rec.count() != 1 {
		t.Fatalf("expected a single batch, got %v", rec.batches)
	}
}

func TestMaxBatchSplitsLargeRequests(t *testing.T) {
	rec := &recorder{}
	l := New(context.Background(), Config{Wait: 10 * time.Millisecond, MaxBatch: 2}, rec.fetch)

	got, err := l.LoadMany(context.Background(), []int{1, 2, 3, 4, 5})
	if err != nil || len(got) != 5 {
		t.Fatalf("got %v, %v", got, err)
	}
	for _, b := range rec.batches {
		if len(b) > 2 {
			t.Fatalf("batch exceeded MaxBatch: %v", b)
		}
	}
}

func TestCancelledCallerStopsWaitingButBatchStillRuns(t *testing.T) {
	release := make(chan struct{})
	l := New(context.Background(), Config{Wait: time.Millisecond}, func(_ context.Context, keys []int) (map[int]string, error) {
		<-release
		return map[int]string{keys[0]: "late"}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := l.Load(ctx, 1); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	close(release)
	got, err := l.Load(context.Background(), 1)
	if err != nil || got != "late" {
		t.Fatalf("another caller should still receive the batch result, got %q, %v", got, err)
	}
}

func TestBatchUsesLoaderContextNotCallerContext(t *testing.T) {
	type key struct{}
	base := context.WithValue(context.Background(), key{}, "request-scoped")
	l := New(base, Config{Wait: time.Millisecond}, func(ctx context.Context, keys []int) (map[int]string, error) {
		return map[int]string{keys[0]: ctx.Value(key{}).(string)}, nil
	})
	got, err := l.Load(context.Background(), 1)
	if err != nil || got != "request-scoped" {
		t.Fatalf("got %q, %v", got, err)
	}
}
