package config

import (
	"time"

	"myim/internal/auth"
	"myim/internal/runtimeconfig"
)

type Config struct {
	Dsn           string        // PostgreSQL连接串
	Addr          string        // HTTP和WebSocket监听地址
	Auth          auth.Config   // 访问令牌配置
	ReadLimit     int64         // 单个WebSocket消息最大字节数
	WriteWait     time.Duration // WebSocket单次写超时
	PongWait      time.Duration // 等待客户端Pong的最长时间
	PingPeriod    time.Duration // 服务端发送Ping的间隔
	SendQueueSize int           // 单连接待发送消息队列大小
}

func New() *Config {
	return &Config{
		Dsn:           runtimeconfig.DatabaseDSN(),
		Addr:          ":8081",
		Auth:          runtimeconfig.Auth(),
		ReadLimit:     32 * 1024,
		WriteWait:     10 * time.Second,
		PongWait:      60 * time.Second,
		PingPeriod:    50 * time.Second,
		SendQueueSize: 64,
	}
}
