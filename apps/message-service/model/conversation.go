package model

const ConversationTypeDirect int16 = 1 // 单聊会话类型

type Conversation struct {
	ConversationID   string `json:"conversation_id"`   // 会话ID
	ConversationType int16  `json:"conversation_type"` // 会话类型
	DirectKey        string `json:"direct_key"`        // 单聊双方排序后的唯一键
	NextSeq          int64  `json:"next_seq"`          // 下一个会话消息序号
}

type ConversationMember struct {
	ConversationID string `json:"conversation_id"` // 会话ID
	UserID         string `json:"user_id"`         // 会话成员用户ID
}

type ConversationSummary struct {
	ConversationID string `json:"conversation_id"` // 会话ID
	PeerUserID     string `json:"peer_user_id"`    // 单聊对方用户ID
	LatestSeq      int64  `json:"latest_seq"`      // 当前最大已分配消息序号
}
