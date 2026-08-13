package config

import (
	"time"
)

type Config struct {
	Dsn          string        // PostgreSQL连接串
	Addr         string        // HTTP监听地址
	JWTSecret    string        // JWT签名密钥
	TokenExpires time.Duration // 访问令牌有效期
}

func New() *Config {
	return &Config{
		Dsn:          "user=postgres password=123456 host=localhost port=5432 dbname=test sslmode=disable",
		Addr:         ":8080",
		JWTSecret:    "myim-development-secret",
		TokenExpires: 24 * time.Hour,
	}
}
