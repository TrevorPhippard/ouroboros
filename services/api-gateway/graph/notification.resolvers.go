package graph

import (
	"context"
	"fmt"

	"api-gateway/graph/model"

	notificationpb "ouroboros/proto/generated/notification"
)

// MarkNotificationRead is the resolver for the markNotificationRead field.
func (r *mutationResolver) MarkNotificationRead(ctx context.Context, id string) (bool, error) {
	if _, err := requireCurrentUserID(ctx); err != nil {
		return false, err
	}
	res, err := r.NotificationClient.MarkAsRead(ctx, &notificationpb.MarkAsReadRequest{
		NotificationId: id,
	})
	if err != nil {
		return false, fmt.Errorf("failed to mark notification as read: %w", err)
	}
	if res == nil {
		return false, fmt.Errorf("notification service returned no response")
	}
	return res.Success, nil
}

// Notifications is the resolver for the notifications field.
func (r *queryResolver) Notifications(
	ctx context.Context,
	userID string,
	limit *int32,
) ([]*model.Notification, error) {
	reqLimit := int32(20)
	if limit != nil {
		reqLimit = *limit
	}

	if reqLimit < 1 || reqLimit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}
	authenticatedUserID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if authenticatedUserID != userID {
		return nil, fmt.Errorf("forbidden: cannot read another user's notifications")
	}

	res, err := r.NotificationClient.GetNotifications(
		ctx,
		&notificationpb.GetNotificationsRequest{
			UserId: userID,
			Limit:  reqLimit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	if res == nil {
		return []*model.Notification{}, nil
	}

	notifications := make([]*model.Notification, 0, len(res.Notifications))

	for _, n := range res.Notifications {
		if n == nil {
			continue
		}

		notifications = append(notifications, &model.Notification{
			ID:        n.Id,
			UserID:    n.UserId,
			Type:      n.Type,
			ActorID:   n.ActorId,
			EntityID:  optionalString(n.EntityId),
			CreatedAt: n.CreatedAt,
			Read:      n.Read,
		})
	}

	return notifications, nil
}

// NotificationReceived is the resolver for the notificationReceived field.
func (r *subscriptionResolver) NotificationReceived(ctx context.Context, userID string) (<-chan *model.Notification, error) {
	authenticatedUserID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if authenticatedUserID != userID {
		return nil, fmt.Errorf("forbidden: cannot subscribe to another user's notifications")
	}
	return nil, fmt.Errorf("subscriptions are not enabled")
}
