package dao

import "time"

type Message struct {
	ID            int64
	Content       string
	SenderUserID  string
	ParentMessage *int64
	ThumbsUp      int64
	Anonymous     bool
	CreatedAt     time.Time
}
