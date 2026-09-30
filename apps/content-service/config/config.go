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
	Dsn                string                   // PostgreSQL连接串
	Addr               string                   // HTTP监听地址
	Auth               auth.Config              // 访问令牌配置
	KafkaBrokers       []string                 // Kafka broker地址列表
	KafkaTopic         string                   // Content事件主题
	OutboxPollInterval time.Duration            // Outbox轮询间隔
	OutboxLease        time.Duration            // 单次事件投递租约
	OutboxRetryBase    time.Duration            // 失败重试初始间隔
	OutboxBatchSize    int32                    // 单次领取事件数量
	OutboxRetention    time.Duration            // 已发布事件保留时间
	HTTP               runtimeconfig.HTTPConfig // HTTP超时和报文头限制
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
		HTTP:               runtimeconfig.DefaultHTTP(),
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
	if err := c.HTTP.Validate(c.Addr, false); err != nil {
		return err
	}
	if c.OutboxPollInterval <= 0 || c.OutboxLease <= 0 || c.OutboxRetryBase <= 0 || c.OutboxBatchSize <= 0 || c.OutboxRetention <= 0 {
		return errors.New("outbox intervals, limits and retention must be positive")
	}
	return nil
}
