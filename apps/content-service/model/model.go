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
	ContentID    string     // 动态ID
	AuthorUserID string     // 作者用户ID
	Text         string     // 动态文本
	MediaURLs    []string   // 图片URL列表
	Status       int32      // 动态状态
	LikeCount    int64      // 点赞数
	CommentCount int64      // 评论数
	CreatedAt    time.Time  // 创建时间
	PublishedAt  *time.Time // 发布时间
	UpdatedAt    time.Time  // 更新时间
}

type Comment struct {
	CommentID    string    // 评论ID
	ContentID    string    // 动态ID
	AuthorUserID string    // 评论作者用户ID
	ParentID     string    // 父评论ID
	Text         string    // 评论文本
	Status       int32     // 评论状态
	CreatedAt    time.Time // 创建时间
}
