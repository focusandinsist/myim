package model

import "time"

const (
	FollowStatusActive    = 1 // 有效关注状态
	FollowStatusCancelled = 2 // 已取消关注状态
)

type Follow struct {
	FollowerUserID string    `json:"follower_user_id"` // 关注发起用户ID
	FolloweeUserID string    `json:"followee_user_id"` // 被关注用户ID
	Status         int32     `json:"status"`           // 关注关系状态
	CreatedAt      time.Time `json:"created_at"`       // 首次关注时间
	UpdatedAt      time.Time `json:"updated_at"`       // 最后更新时间
}
