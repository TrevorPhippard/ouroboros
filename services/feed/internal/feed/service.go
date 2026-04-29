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

func (s *Service) GetFeed(ctx context.Context, req *pbFeed.GetFeedRequest) (*pbFeed.GetFeedResponse, error) {
	items, next, err := s.store.GetFeed(ctx, req.UserId, req.Cursor, int64(req.Limit))
	if err != nil {
		return nil, err
	}

	return &pbFeed.GetFeedResponse{
		Items:      items,
		NextCursor: next,
	}, nil
}
