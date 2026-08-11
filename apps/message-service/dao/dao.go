package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/message-service/config"

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
		CREATE TABLE IF NOT EXISTS messages (
			message_id VARCHAR(36) PRIMARY KEY,
			request_id VARCHAR(128) NOT NULL,
			conversation_id VARCHAR(73) NOT NULL,
			sender_user_id VARCHAR(36) NOT NULL,
			target_user_id VARCHAR(36) NOT NULL,
			message_type SMALLINT NOT NULL,
			content TEXT NOT NULL,
			sent_at BIGINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT messages_sender_request_unique UNIQUE (sender_user_id, request_id)
		);

		CREATE INDEX IF NOT EXISTS messages_conversation_sent_at_idx
		ON messages (conversation_id, sent_at DESC, message_id DESC);

		CREATE INDEX IF NOT EXISTS messages_target_sent_at_idx
		ON messages (target_user_id, sent_at DESC, message_id DESC);
	`
	if _, err := d.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("initialize messages schema: %w", err)
	}
	return nil
}

func (d *Dao) Close() error {
	return d.db.Close()
}
