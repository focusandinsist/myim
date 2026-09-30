package model

type OutboxEvent struct {
	OutboxID     int64
	EventID      string
	EventType    string
	AggregateID  string
	Topic        string
	PartitionKey string
	Payload      []byte
	Attempts     int32
	ClaimToken   string
}
