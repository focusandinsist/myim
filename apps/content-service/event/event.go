package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
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
	Publish(topic, key string, value []byte) error // 发布一个事件
}

type SaramaPublisher struct {
	producer sarama.SyncProducer // Sarama同步生产者
}

func NewSaramaPublisher(brokers []string) (*SaramaPublisher, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	config.Net.DialTimeout = 5 * time.Second
	config.Net.ReadTimeout = 5 * time.Second
	config.Net.WriteTimeout = 5 * time.Second
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}
	return &SaramaPublisher{producer: producer}, nil
}

func (p *SaramaPublisher) Publish(topic, key string, value []byte) error {
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(value)})
	return err
}

func (p *SaramaPublisher) Close() error {
	return p.producer.Close()
}

func LikeEnvelope(eventType, contentID, userID string, liked bool) (Envelope, error) {
	payload, err := json.Marshal(map[string]any{
		"content_id": contentID,
		"user_id":    userID,
		"liked":      liked,
	})
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal like event payload: %w", err)
	}
	return Envelope{
		EventID:      uuid.NewString(),
		EventType:    eventType,
		EventVersion: 1,
		AggregateID:  contentID,
		OccurredAt:   time.Now().UTC(),
		Producer:     "content-service",
		Payload:      payload,
	}, nil
}
