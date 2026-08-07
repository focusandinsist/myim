package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/app/model"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserNotFound   = errors.New("user not found")           // 按指定条件没有查询到用户
	ErrUserNameExists = errors.New("user name already exists") // 用户名违反数据库唯一约束
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error                        // 创建用户
	GetUserByUserName(ctx context.Context, userName string) (*model.User, error)   // 根据用户名查询用户
	UserNameExists(ctx context.Context, userName string) (bool, error)             // 判断用户名是否已经存在
	UpdateLastLoginTime(ctx context.Context, userID string, loginTime int64) error // 更新用户最后登录时间
}

func (d *Dao) CreateUser(ctx context.Context, user *model.User) error {
	const query = `
		INSERT INTO users
			(user_id, user_name, password, phone, email, nickname, avatar, bio,
			 gender, birthday, region, status, register_time, last_login_time, updated_time)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := d.db.ExecContext(
		ctx,
		query,
		user.UserID,
		user.UserName,
		user.Password,
		user.Phone,
		user.Email,
		user.Nickname,
		user.Avatar,
		user.Bio,
		user.Gender,
		user.Birthday,
		user.Region,
		user.Status,
		user.RegisterTime,
		user.LastLoginTime,
		user.UpdatedTime,
	)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrUserNameExists
	}
	return fmt.Errorf("create user: %w", err)
}

func (d *Dao) GetUserByUserID(ctx context.Context, userID string) (*model.User, error) {
	const query = `
		SELECT
			user_id, user_name, password, phone, email, nickname, avatar, bio,
			gender, birthday, region, status, register_time, last_login_time, updated_time
		FROM users
		WHERE user_id = $1
	`
	return d.scanUser(d.db.QueryRowContext(ctx, query, userID))
}

func (d *Dao) GetUserByUserName(ctx context.Context, userName string) (*model.User, error) {
	const query = `
		SELECT
			user_id, user_name, password, phone, email, nickname, avatar, bio,
			gender, birthday, region, status, register_time, last_login_time, updated_time
		FROM users
		WHERE LOWER(user_name) = LOWER($1)
	`
	return d.scanUser(d.db.QueryRowContext(ctx, query, userName))
}

func (d *Dao) UserNameExists(ctx context.Context, userName string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE LOWER(user_name) = LOWER($1)
		)
	`
	var exists bool
	if err := d.db.QueryRowContext(ctx, query, userName).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user name exists: %w", err)
	}
	return exists, nil
}

func (d *Dao) UpdateLastLoginTime(ctx context.Context, userID string, loginTime int64) error {
	const query = `
		UPDATE users
		SET last_login_time = $1
		WHERE user_id = $2
	`
	result, err := d.db.ExecContext(ctx, query, loginTime, userID)
	if err != nil {
		return fmt.Errorf("update last login time: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated rows: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (d *Dao) scanUser(row rowScanner) (*model.User, error) {
	user := new(model.User)
	if err := row.Scan(
		&user.UserID,
		&user.UserName,
		&user.Password,
		&user.Phone,
		&user.Email,
		&user.Nickname,
		&user.Avatar,
		&user.Bio,
		&user.Gender,
		&user.Birthday,
		&user.Region,
		&user.Status,
		&user.RegisterTime,
		&user.LastLoginTime,
		&user.UpdatedTime,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return user, nil
}
