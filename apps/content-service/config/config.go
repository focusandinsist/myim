package config

import "os"

type Config struct {
	Dsn          string   // PostgreSQL连接串
	Addr         string   // HTTP监听地址
	JWTSecret    string   // JWT签名密钥
	KafkaBrokers []string // Kafka broker地址列表
	KafkaTopic   string   // Content事件主题
}

func New() *Config {
	brokers := []string{}
	if value := os.Getenv("MYIM_KAFKA_BROKERS"); value != "" {
		brokers = []string{value}
	}
	return &Config{Dsn: "user=postgres password=123456 host=localhost port=5432 dbname=test sslmode=disable", Addr: ":8082", JWTSecret: "myim-development-secret", KafkaBrokers: brokers, KafkaTopic: "content-events"}
}
