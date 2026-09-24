package dao

import (
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/content-service/config"
	contentdb "myim/internal/db/content"

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
	return &Dao{db: db, queries: contentdb.New(db)}, nil
}

func (d *Dao) Close() error { return d.db.Close() }
