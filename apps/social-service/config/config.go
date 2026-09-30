package config

import (
	"myim/internal/auth"
	"myim/internal/runtimeconfig"
)

type Config struct {
	Dsn  string      // PostgreSQL连接串
	Addr string      // HTTP监听地址
	Auth auth.Config // 访问令牌配置
}

func New() *Config {
	return &Config{
		Dsn:  runtimeconfig.DatabaseDSN(),
		Addr: ":8083",
		Auth: runtimeconfig.Auth(),
	}
}
