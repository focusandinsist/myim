package config

import "time"

type Config struct {
	Addr          string        // HTTP和WebSocket监听地址
	JWTSecret     string        // JWT签名密钥
	ReadLimit     int64         // 单个WebSocket消息最大字节数
	WriteWait     time.Duration // WebSocket单次写超时
	PongWait      time.Duration // 等待客户端Pong的最长时间
	PingPeriod    time.Duration // 服务端发送Ping的间隔
	SendQueueSize int           // 单连接待发送消息队列大小
}

func New() *Config {
	return &Config{
		Addr:          ":8081",
		JWTSecret:     "myim-development-secret",
		ReadLimit:     32 * 1024,
		WriteWait:     10 * time.Second,
		PongWait:      60 * time.Second,
		PingPeriod:    50 * time.Second,
		SendQueueSize: 64,
	}
}
