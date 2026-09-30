package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"myim/internal/auth"
	"myim/internal/runtimeconfig"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	Dsn           string                   // PostgreSQL连接串
	Addr          string                   // HTTP和WebSocket监听地址
	Auth          auth.Config              // 访问令牌配置
	ReadLimit     int64                    // 单个WebSocket消息最大字节数
	WriteWait     time.Duration            // WebSocket单次写超时
	PongWait      time.Duration            // 等待客户端Pong的最长时间
	PingPeriod    time.Duration            // 服务端发送Ping的间隔
	SendQueueSize int                      // 单连接待发送消息队列大小
	HTTP          runtimeconfig.HTTPConfig // HTTP超时和报文头限制
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
		HTTP:          runtimeconfig.DefaultHTTP(),
	}
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Dsn) == "" {
		return errors.New("database DSN is required")
	}
	if _, err := pgx.ParseConfig(c.Dsn); err != nil {
		return fmt.Errorf("invalid database DSN: %w", err)
	}
	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("validate auth config: %w", err)
	}
	if err := c.HTTP.Validate(c.Addr, true); err != nil {
		return err
	}
	if c.ReadLimit <= 0 || c.WriteWait <= 0 || c.PongWait <= 0 || c.PingPeriod <= 0 || c.PingPeriod >= c.PongWait || c.SendQueueSize <= 0 {
		return errors.New("websocket limits and intervals are invalid")
	}
	return nil
}
