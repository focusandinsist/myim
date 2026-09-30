package config

import (
	"myim/internal/auth"
	"myim/internal/runtimeconfig"
)

type Config struct {
	Dsn          string      // PostgreSQL连接串
	Addr         string      // HTTP监听地址
	Auth         auth.Config // 访问令牌配置
	KafkaBrokers []string    // Kafka broker地址列表
	KafkaTopic   string      // Content事件主题
}

func New() *Config {
	return &Config{
		Dsn:          runtimeconfig.DatabaseDSN(),
		Addr:         ":8082",
		Auth:         runtimeconfig.Auth(),
		KafkaBrokers: []string{"127.0.0.1:9092"},
		KafkaTopic:   "content-events",
	}
}
