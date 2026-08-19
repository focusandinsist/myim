package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/social-service/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Dao struct {
	db *sql.DB // PostgreSQL连接池
}

func New(c *config.Config) (*Dao, error) {
	if c.Dsn == "" {
		return nil, errors.New("empty dsn")
	}
	db, err := sql.Open("pgx", c.Dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	dao := &Dao{db: db}
	if err := dao.initSchema(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return dao, nil
}

func (d *Dao) initSchema(ctx context.Context) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS follows (
			follower_user_id VARCHAR(36) NOT NULL,
			followee_user_id VARCHAR(36) NOT NULL,
			status SMALLINT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_user_id, followee_user_id),
			CHECK (follower_user_id <> followee_user_id)
		);

		CREATE INDEX IF NOT EXISTS follows_followee_status_idx
		ON follows (followee_user_id, status, updated_at DESC);

		CREATE INDEX IF NOT EXISTS follows_follower_status_idx
		ON follows (follower_user_id, status, updated_at DESC);
	`
	if _, err := d.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("initialize social schema: %w", err)
	}
	return nil
}

func (d *Dao) Close() error {
	return d.db.Close()
}
