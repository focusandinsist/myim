package model

const ConversationTypeDirect int16 = 1 // 单聊会话类型

type Conversation struct {
	ConversationID   string `json:"conversation_id"   gorm:"column:conversation_id;primaryKey;type:varchar(36)"` // 会话ID
	ConversationType int16  `json:"conversation_type" gorm:"column:conversation_type;not null"`                  // 会话类型
	DirectKey        string `json:"direct_key"        gorm:"column:direct_key;type:varchar(73);uniqueIndex"`     // 单聊双方排序后的唯一键
	NextSeq          int64  `json:"next_seq"          gorm:"column:next_seq;not null"`                           // 下一个会话消息序号
}

func (Conversation) TableName() string {
	return "conversations"
}

type ConversationMember struct {
	ConversationID string `json:"conversation_id" gorm:"column:conversation_id;primaryKey;type:varchar(36)"` // 会话ID
	UserID         string `json:"user_id"         gorm:"column:user_id;primaryKey;type:varchar(36)"`         // 会话成员用户ID
}

func (ConversationMember) TableName() string {
	return "conversation_members"
}
