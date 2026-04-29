package feed

import (
	"context"
	"fmt"

	pbFeed "ouroboros/proto/generated/feed"

	"github.com/go-redis/redis/v8"
)

func (s *Store) GetFeed(ctx context.Context, userID, cursor string, limit int64) ([]*pbFeed.FeedItem, string, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	maxScore := "+inf"
	if cursor != "" {
		maxScore = cursor
	}

	results, err := s.rdb.ZRevRangeByScoreWithScores(ctx, feedKey(userID), &redis.ZRangeBy{
		Max:   maxScore,
		Min:   "-inf",
		Count: limit,
	}).Result()

	if err != nil {
		return nil, "", err
	}

	items := make([]*pbFeed.FeedItem, 0, len(results))
	var lastScore float64

	for _, r := range results {
		postID, _ := r.Member.(string)

		items = append(items, &pbFeed.FeedItem{
			PostId: postID,
		})

		lastScore = r.Score
	}

	next := ""
	if len(items) == int(limit) {
		next = fmt.Sprintf("%.0f", lastScore-1)
	}

	return items, next, nil
}
