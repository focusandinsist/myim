package event

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const TopicContentEvents = "content-events" // Content业务事件主题

type Envelope struct {
	EventID      string          `json:"event_id"`      // 事件唯一ID
	EventType    string          `json:"event_type"`    // 事件类型
	EventVersion int             `json:"event_version"` // 事件版本
	AggregateID  string          `json:"aggregate_id"`  // 聚合根ID
	OccurredAt   time.Time       `json:"occurred_at"`   // 业务发生时间
	Producer     string          `json:"producer"`      // 事件生产者
	Payload      json.RawMessage `json:"payload"`       // 业务载荷
}

type Publisher interface {
	Publish(topic, key string, value []byte) error
} // 事件发布接口

type LogPublisher struct{} // 未配置Kafka时的本地事件发布器

func (LogPublisher) Publish(topic, key string, value []byte) error {
	log.Printf("content event topic=%s key=%s payload=%s", topic, key, value)
	return nil
}

func LikeEnvelope(eventType, contentID, userID string, liked bool) ([]byte, error) {
	payload, err := json.Marshal(map[string]any{"content_id": contentID, "user_id": userID, "liked": liked})
	if err != nil {
		return nil, fmt.Errorf("marshal like event payload: %w", err)
	}
	return json.Marshal(Envelope{EventID: fmt.Sprintf("%s-%s-%s", eventType, contentID, userID), EventType: eventType, EventVersion: 1, AggregateID: contentID, OccurredAt: time.Now().UTC(), Producer: "content-service", Payload: payload})
}
