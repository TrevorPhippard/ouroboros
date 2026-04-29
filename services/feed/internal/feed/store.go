package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pbFeed "ouroboros/proto/generated/feed"

	"github.com/go-redis/redis/v8"
)

func feedKey(userID string) string {
	return fmt.Sprintf("feed:%s", userID)
}

func jobKey(eventID string) string {
	return "fanout:job:" + eventID
}

func parseTimestamp(ts string) (float64, error) {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return 0, err
	}
	return float64(t.UnixNano()), nil
}

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) FanoutBatch(ctx context.Context, userIDs []string, item *pbFeed.FeedItem) error {
	score, err := parseTimestamp(item.Post.Timestamp)
	if err != nil {
		return err
	}

	pipe := s.rdb.Pipeline()

	for _, userID := range userIDs {
		key := feedKey(userID)

		pipe.ZAdd(ctx, key, &redis.Z{
			Score:  score,
			Member: item.PostId,
		})

		pipe.ZRemRangeByRank(ctx, key, 0, -(1000 + 1))
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) SaveJob(ctx context.Context, job *FanoutJob) error {
	data, _ := json.Marshal(job)
	return s.rdb.Set(ctx, jobKey(job.EventID), data, 24*time.Hour).Err()
}

func (s *Store) LoadJob(ctx context.Context, eventID string) (*FanoutJob, error) {
	data, err := s.rdb.Get(ctx, jobKey(eventID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var job FanoutJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}

	return &job, nil
}
