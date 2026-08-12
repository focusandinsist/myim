package model

type Message struct {
	MessageID      string `json:"message_id"      gorm:"column:message_id;primaryKey;type:varchar(36)"`          // 服务端生成的唯一消息ID
	RequestID      string `json:"request_id"      gorm:"column:request_id;type:varchar(128);not null"`           // 客户端生成的幂等请求ID
	ConversationID string `json:"conversation_id" gorm:"column:conversation_id;type:varchar(36);not null;index"` // 消息所属会话ID
	SenderUserID   string `json:"sender_user_id"  gorm:"column:sender_user_id;type:varchar(36);not null;index"`  // 发送方用户ID
	TargetUserID   string `json:"target_user_id"  gorm:"column:target_user_id;type:varchar(36);not null;index"`  // 接收方用户ID
	MessageType    int32  `json:"message_type"    gorm:"column:message_type;type:smallint;not null"`             // 消息类型
	Content        string `json:"content"         gorm:"column:content;type:text;not null"`                      // 消息内容
	SentAt         int64  `json:"sent_at"         gorm:"column:sent_at;not null;index"`                          // 服务端接收时间，Unix毫秒
	Seq            int64  `json:"seq"             gorm:"column:seq;not null"`                                    // 会话内递增消息序号
}

func (Message) TableName() string {
	return "messages"
}
