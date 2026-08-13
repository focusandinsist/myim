package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/user-service/config"

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
		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(36) PRIMARY KEY,
			user_name VARCHAR(32) NOT NULL,
			password VARCHAR(60) NOT NULL,
			phone VARCHAR(32) NOT NULL DEFAULT '',
			email VARCHAR(128) NOT NULL DEFAULT '',
			nickname VARCHAR(64) NOT NULL DEFAULT '',
			avatar VARCHAR(512) NOT NULL DEFAULT '',
			bio VARCHAR(256) NOT NULL DEFAULT '',
			gender SMALLINT NOT NULL DEFAULT 0,
			birthday BIGINT NOT NULL DEFAULT 0,
			region VARCHAR(128) NOT NULL DEFAULT '',
			status SMALLINT NOT NULL DEFAULT 0,
			register_time BIGINT NOT NULL,
			last_login_time BIGINT NOT NULL DEFAULT 0,
			updated_time BIGINT NOT NULL DEFAULT 0
		);

		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS phone VARCHAR(32) NOT NULL DEFAULT '';
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS email VARCHAR(128) NOT NULL DEFAULT '';
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS nickname VARCHAR(64) NOT NULL DEFAULT '';
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS avatar VARCHAR(512) NOT NULL DEFAULT '';
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS bio VARCHAR(256) NOT NULL DEFAULT '';
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS gender SMALLINT NOT NULL DEFAULT 0;
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS birthday BIGINT NOT NULL DEFAULT 0;
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS region VARCHAR(128) NOT NULL DEFAULT '';
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS status SMALLINT NOT NULL DEFAULT 0;
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS updated_time BIGINT NOT NULL DEFAULT 0;

		CREATE UNIQUE INDEX IF NOT EXISTS users_user_name_lower_idx
		ON users (LOWER(user_name));

		CREATE UNIQUE INDEX IF NOT EXISTS users_phone_unique_idx
		ON users (phone)
		WHERE phone <> '';

		CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_unique_idx
		ON users (LOWER(email))
		WHERE email <> '';
	`
	if _, err := d.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("initialize users schema: %w", err)
	}
	return nil
}

func (d *Dao) Close() error {
	return d.db.Close()
}
