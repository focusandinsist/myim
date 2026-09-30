package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"myim/apps/content-service/config"
	contentdb "myim/db/content"
	"myim/internal/migration"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Dao struct {
	db      *sql.DB
	queries *contentdb.Queries
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
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres after migrations: %w", err)
	}
	return &Dao{db: db, queries: contentdb.New(db)}, nil
}

func (d *Dao) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *Dao) Close() error { return d.db.Close() }
