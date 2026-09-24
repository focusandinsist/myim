package model

type Message struct {
	MessageID      string `json:"message_id"`      // 服务端生成的唯一消息ID
	RequestID      string `json:"request_id"`      // 客户端生成的幂等请求ID
	ConversationID string `json:"conversation_id"` // 消息所属会话ID
	SenderUserID   string `json:"sender_user_id"`  // 发送方用户ID
	TargetUserID   string `json:"target_user_id"`  // 接收方用户ID
	MessageType    int32  `json:"message_type"`    // 消息类型
	Content        string `json:"content"`         // 消息内容
	SentAt         int64  `json:"sent_at"`         // 服务端接收时间，Unix毫秒
	Seq            int64  `json:"seq"`             // 会话内递增消息序号
}
