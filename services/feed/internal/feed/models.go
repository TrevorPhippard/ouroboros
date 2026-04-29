package feed

type PostCreatedEvent struct {
	EventID   string `json:"eventId"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		PostID   string `json:"postId"`
		AuthorID string `json:"authorId"`
		Content  string `json:"content"`
	} `json:"data"`
}

type FanoutJob struct {
	EventID   string
	PostID    string
	AuthorID  string
	Followers []string
	BatchSize int
	Completed map[int]bool
	Cursor    int
}
