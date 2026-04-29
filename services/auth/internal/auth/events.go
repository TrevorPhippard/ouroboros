package auth

import "time"

type userSignedUpEvent struct {
	EventID   string `json:"eventId"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		UserID      string `json:"userId"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	} `json:"data"`
}

func newUserSignedUpEvent(userID, email, displayName string) userSignedUpEvent {
	event := userSignedUpEvent{
		EventID:   "signup-" + userID,
		Type:      "UserSignedUp",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	event.Data.UserID = userID
	event.Data.Email = email
	event.Data.DisplayName = displayName
	return event
}
