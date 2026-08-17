package config

type Config struct {
	Dsn          string   // PostgreSQL连接串
	Addr         string   // HTTP监听地址
	JWTSecret    string   // JWT签名密钥
	KafkaBrokers []string // Kafka broker地址列表
	KafkaTopic   string   // Content事件主题
}

func New() *Config {
	return &Config{
		Dsn:          "user=postgres password=123456 host=localhost port=5432 dbname=test sslmode=disable",
		Addr:         ":8082",
		JWTSecret:    "myim-development-secret",
		KafkaBrokers: []string{"127.0.0.1:9092"},
		KafkaTopic:   "content-events",
	}
}
