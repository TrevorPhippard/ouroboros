package feed

import (
	"context"

	pbFeed "ouroboros/proto/generated/feed"
)

type Service struct {
	pbFeed.UnimplementedFeedServiceServer
	store  *Store
	social *SocialGraph
	fanout *FanoutEngine
}

func NewService(store *Store, social *SocialGraph, fanout *FanoutEngine) *Service {
	return &Service{
		store:  store,
		social: social,
		fanout: fanout,
	}
}

func (s *Service) HandlePostCreated(ctx context.Context, event PostCreatedEvent) error {
	item := &pbFeed.FeedItem{
		PostId: event.Data.PostID,
		Post: &pbFeed.Post{
			Id:        event.Data.PostID,
			AuthorId:  event.Data.AuthorID,
			Content:   event.Data.Content,
			Timestamp: event.Timestamp,
		},
	}

	job, err := s.store.LoadJob(ctx, event.EventID)
	if err != nil {
		return err
	}

	if job == nil {
		followers, _ := s.social.GetFollowers(ctx, event.Data.AuthorID)

		job = &FanoutJob{
			EventID:   event.EventID,
			PostID:    event.Data.PostID,
			AuthorID:  event.Data.AuthorID,
			Followers: followers,
			BatchSize: 100,
			Completed: make(map[int]bool),
		}
	}

	return s.fanout.Run(ctx, s.store, job, item)
}

func (s *Service) GetFeed(
	ctx context.Context,
	req *pbFeed.GetFeedRequest,
) (*pbFeed.GetFeedResponse, error) {
	// Dummy feed data for development/testing.
	items := []*pbFeed.FeedItem{
		{
			PostId: "post-1",
			Cursor: "cursor-001",
		},
		{
			PostId: "post-2",
			Cursor: "cursor-002",
		},
		{
			PostId: "post-3",
			Cursor: "cursor-003",
		},
	}

	// Basic cursor pagination.
	start := 0
	if req.Cursor != "" {
		for i, item := range items {
			if item.Cursor == req.Cursor {
				start = i + 1
				break
			}
		}
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	if start >= len(items) {
		return &pbFeed.GetFeedResponse{
			Items:      []*pbFeed.FeedItem{},
			NextCursor: "",
		}, nil
	}

	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	page := items[start:end]

	nextCursor := ""
	if end < len(items) {
		nextCursor = items[end-1].Cursor
	}

	return &pbFeed.GetFeedResponse{
		Items:      page,
		NextCursor: nextCursor,
	}, nil
}
