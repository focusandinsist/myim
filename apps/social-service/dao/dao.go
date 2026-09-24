package dao

import (
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/social-service/config"
	socialdb "myim/db/social"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Dao struct {
	db      *sql.DB
	queries *socialdb.Queries
}

func New(c *config.Config) (*Dao, error) {
	if c.Dsn == "" {
		return nil, errors.New("empty dsn")
	}
	db, err := sql.Open("pgx", c.Dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return &Dao{db: db, queries: socialdb.New(db)}, nil
}

func (d *Dao) Close() error {
	return d.db.Close()
}
