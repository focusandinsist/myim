package config

import (
	"errors"
	"fmt"
	"strings"

	"myim/internal/auth"
	"myim/internal/runtimeconfig"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	Dsn  string                   // PostgreSQL连接串
	Addr string                   // HTTP监听地址
	Auth auth.Config              // 访问令牌配置
	HTTP runtimeconfig.HTTPConfig // HTTP超时和报文头限制
}

func New() *Config {
	return &Config{
		Dsn:  runtimeconfig.DatabaseDSN(),
		Addr: ":8080",
		Auth: runtimeconfig.Auth(),
		HTTP: runtimeconfig.DefaultHTTP(),
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
	return c.HTTP.Validate(c.Addr, false)
}
