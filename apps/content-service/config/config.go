package config

import (
	"time"

	"myim/internal/auth"
	"myim/internal/runtimeconfig"
)

type Config struct {
	Dsn                string        // PostgreSQL连接串
	Addr               string        // HTTP监听地址
	Auth               auth.Config   // 访问令牌配置
	KafkaBrokers       []string      // Kafka broker地址列表
	KafkaTopic         string        // Content事件主题
	OutboxPollInterval time.Duration // Outbox轮询间隔
	OutboxLease        time.Duration // 单次事件投递租约
	OutboxRetryBase    time.Duration // 失败重试初始间隔
	OutboxBatchSize    int32         // 单次领取事件数量
	OutboxRetention    time.Duration // 已发布事件保留时间
}

func New() *Config {
	return &Config{
		Dsn:                runtimeconfig.DatabaseDSN(),
		Addr:               ":8082",
		Auth:               runtimeconfig.Auth(),
		KafkaBrokers:       []string{"127.0.0.1:9092"},
		KafkaTopic:         "content-events",
		OutboxPollInterval: time.Second,
		OutboxLease:        2 * time.Minute,
		OutboxRetryBase:    time.Second,
		OutboxBatchSize:    32,
		OutboxRetention:    7 * 24 * time.Hour,
	}
}
