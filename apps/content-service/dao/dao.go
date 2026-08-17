package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/content-service/config"

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
		_ = db.Close()
		return nil, err
	}
	return dao, nil
}

func (d *Dao) initSchema(ctx context.Context) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS contents (
			content_id VARCHAR(36) PRIMARY KEY,
			author_user_id VARCHAR(36) NOT NULL,
			text TEXT NOT NULL,
			media_urls JSONB NOT NULL DEFAULT '[]',
			status SMALLINT NOT NULL,
			like_count BIGINT NOT NULL DEFAULT 0,
			comment_count BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL,
			published_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ NOT NULL
		);

		CREATE INDEX IF NOT EXISTS contents_author_status_idx
		ON contents (author_user_id, status, created_at DESC);

		CREATE TABLE IF NOT EXISTS content_likes (
			content_id VARCHAR(36) NOT NULL REFERENCES contents(content_id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (content_id, user_id)
		);

		CREATE TABLE IF NOT EXISTS content_comments (
			comment_id VARCHAR(36) PRIMARY KEY,
			content_id VARCHAR(36) NOT NULL REFERENCES contents(content_id) ON DELETE CASCADE,
			author_user_id VARCHAR(36) NOT NULL,
			parent_id VARCHAR(36) NOT NULL DEFAULT '',
			text TEXT NOT NULL,
			status SMALLINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS content_comments_content_idx
		ON content_comments (content_id, created_at DESC);

		CREATE TABLE IF NOT EXISTS follows (
			follower_user_id VARCHAR(36) NOT NULL,
			followee_user_id VARCHAR(36) NOT NULL,
			status SMALLINT NOT NULL DEFAULT 1,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_user_id, followee_user_id),
			CHECK (follower_user_id <> followee_user_id)
		);`
	if _, err := d.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("initialize content schema: %w", err)
	}
	return nil
}

func (d *Dao) Close() error { return d.db.Close() }
