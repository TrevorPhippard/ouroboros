package feed

type PostCreatedEvent struct {
	EventID   string `json:"eventId"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		PostID  string `json:"postId"`
		UserId  string `json:"UserId"`
		Content string `json:"content"`
	} `json:"data"`
}

type FanoutJob struct {
	EventID   string
	PostID    string
	UserId    string
	Followers []string
	BatchSize int
	Completed map[int]bool
	Cursor    int
}
