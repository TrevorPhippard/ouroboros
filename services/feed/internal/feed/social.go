package feed

import "context"

type SocialGraph struct{}

func NewSocialGraph() *SocialGraph {
	return &SocialGraph{}
}

func (g *SocialGraph) GetFollowers(ctx context.Context, userID string) ([]string, error) {
	return []string{"user-1", "user-2", "user-3", userID}, nil
}
