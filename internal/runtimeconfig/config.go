package runtimeconfig

import (
	"os"
	"time"

	"myim/internal/auth"
)

const DevelopmentDatabaseDSN = "user=postgres password=123456 host=localhost port=5432 dbname=test sslmode=disable"

func DatabaseDSN() string {
	if dsn := os.Getenv("MYIM_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	return DevelopmentDatabaseDSN
}

func Auth() auth.Config {
	secret := os.Getenv("MYIM_JWT_SECRET")
	if secret == "" {
		secret = auth.DevelopmentSecret
	}
	return auth.Config{Secret: secret, TokenExpires: 24 * time.Hour}
}
