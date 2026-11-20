// Package mapping converts protobuf messages from the backend services into the
// GraphQL models. It is the one place that knows both vocabularies.
package mapping

import (
	"api-gateway/graph/model"

	notificationpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

// OptionalString maps "" to nil, for nullable GraphQL fields.
func OptionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func Post(p *postpb.Post) *model.Post {
	if p == nil {
		return nil
	}
	return &model.Post{
		ID:        p.Id,
		Content:   p.Content,
		CreatedAt: p.CreatedAt,
		UserID:    p.AuthorId,
	}
}

func Comment(c *postpb.Comment) *model.Comment {
	if c == nil {
		return nil
	}
	return &model.Comment{
		ID:        c.Id,
		PostID:    c.PostId,
		UserID:    c.AuthorId,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
	}
}

func Notification(n *notificationpb.Notification) *model.Notification {
	if n == nil {
		return nil
	}
	return &model.Notification{
		ID:        n.Id,
		UserID:    n.UserId,
		Type:      n.Type,
		ActorID:   n.ActorId,
		EntityID:  OptionalString(n.EntityId),
		CreatedAt: n.CreatedAt,
		Read:      n.Read,
	}
}

func Profile(p *profilepb.Profile) *model.Profile {
	if p == nil {
		return nil
	}
	experiences := make([]*model.Experience, 0, len(p.Experiences))
	for _, e := range p.Experiences {
		if e == nil {
			continue
		}
		experiences = append(experiences, &model.Experience{
			ID:          e.Id,
			Title:       e.Title,
			Company:     e.Company,
			StartDate:   e.StartDate,
			EndDate:     OptionalString(e.EndDate),
			Description: e.Description,
		})
	}
	return &model.Profile{
		ID:          p.Id,
		UserID:      p.UserId,
		Headline:    OptionalString(p.Headline),
		About:       OptionalString(p.About),
		DisplayName: OptionalString(p.DisplayName),
		AvatarURL:   OptionalString(p.AvatarUrl),
		Bio:         OptionalString(p.Bio),
		Experiences: experiences,
	}
}
