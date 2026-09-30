package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"myim/apps/user-service/config"
	userdb "myim/db/user"
	"myim/internal/migration"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Dao struct {
	db      *sql.DB
	queries *userdb.Queries
}

func New(c *config.Config) (*Dao, error) {
	if c.Dsn == "" {
		return nil, errors.New("empty dsn")
	}
	db, err := sql.Open("pgx", c.Dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := migration.Up(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run database migrations: %w", err)
	}
	return &Dao{db: db, queries: userdb.New(db)}, nil
}

func (d *Dao) Close() error {
	return d.db.Close()
}
