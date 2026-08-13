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
		CREATE TABLE IF NOT EXISTS conversations (
			conversation_id VARCHAR(36) PRIMARY KEY,
			conversation_type SMALLINT NOT NULL,
			direct_key VARCHAR(73) UNIQUE,
			next_seq BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT conversations_direct_key_required CHECK (
				(conversation_type = 1 AND direct_key IS NOT NULL) OR
				(conversation_type <> 1 AND direct_key IS NULL)
			)
		);

		CREATE TABLE IF NOT EXISTS conversation_members (
			conversation_id VARCHAR(36) NOT NULL REFERENCES conversations(conversation_id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL,
			joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (conversation_id, user_id)
		);

		CREATE INDEX IF NOT EXISTS conversation_members_user_id_idx
		ON conversation_members (user_id, conversation_id);

		CREATE TABLE IF NOT EXISTS messages (
			message_id VARCHAR(36) PRIMARY KEY,
			request_id VARCHAR(128) NOT NULL,
			conversation_id VARCHAR(36) NOT NULL,
			sender_user_id VARCHAR(36) NOT NULL,
			target_user_id VARCHAR(36) NOT NULL,
			message_type SMALLINT NOT NULL,
			content TEXT NOT NULL,
			sent_at BIGINT NOT NULL,
			seq BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT messages_sender_request_unique UNIQUE (sender_user_id, request_id)
		);

		ALTER TABLE messages
		ADD COLUMN IF NOT EXISTS seq BIGINT NOT NULL DEFAULT 0;

		CREATE UNIQUE INDEX IF NOT EXISTS messages_conversation_seq_unique
		ON messages (conversation_id, seq)
		WHERE seq > 0;

		CREATE INDEX IF NOT EXISTS messages_conversation_sent_at_idx
		ON messages (conversation_id, seq DESC);

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
