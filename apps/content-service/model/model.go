package model

import "time"

const (
	StatusDraft     = 1
	StatusPublished = 2
	StatusDeleted   = 3
	CommentActive   = 1
	CommentDeleted  = 2
)

type Content struct {
	ContentID    string
	AuthorUserID string
	Text         string
	MediaURLs    []string
	Status       int32
	LikeCount    int64
	CommentCount int64
	CreatedAt    time.Time
	PublishedAt  *time.Time
	UpdatedAt    time.Time
}

type Comment struct {
	CommentID    string
	ContentID    string
	AuthorUserID string
	ParentID     string
	Text         string
	Status       int32
	CreatedAt    time.Time
}
