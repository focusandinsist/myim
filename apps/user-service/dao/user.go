package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/user-service/model"
	userdb "myim/internal/db/user"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserNotFound   = errors.New("user not found")           // 按指定条件没有查询到用户
	ErrUserNameExists = errors.New("user name already exists") // 用户名违反数据库唯一约束
)

func (d *Dao) CreateUser(ctx context.Context, user *model.User) error {
	err := d.queries.CreateUser(ctx, userdb.CreateUserParams{UserID: user.UserID, UserName: user.UserName, Password: user.Password, Phone: user.Phone, Email: user.Email, Nickname: user.Nickname, Avatar: user.Avatar, Bio: user.Bio, Gender: int16(user.Gender), Birthday: user.Birthday, Region: user.Region, Status: int16(user.Status), RegisterTime: user.RegisterTime, LastLoginTime: user.LastLoginTime, UpdatedTime: user.UpdatedTime})
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
	row, err := d.queries.GetUserByUserID(ctx, userID)
	return d.convertUser(row, err)
}

func (d *Dao) GetUserByUserName(ctx context.Context, userName string) (*model.User, error) {
	row, err := d.queries.GetUserByUserName(ctx, userName)
	return d.convertUser(row, err)
}

func (d *Dao) UserNameExists(ctx context.Context, userName string) (bool, error) {
	exists, err := d.queries.UserNameExists(ctx, userName)
	if err != nil {
		return false, fmt.Errorf("check user name exists: %w", err)
	}
	return exists, nil
}

func (d *Dao) UpdateLastLoginTime(ctx context.Context, userID string, loginTime int64) error {
	rows, err := d.queries.UpdateLastLoginTime(ctx, userdb.UpdateLastLoginTimeParams{LastLoginTime: loginTime, UserID: userID})
	if err != nil {
		return fmt.Errorf("update last login time: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (d *Dao) convertUser(row userdb.User, err error) (*model.User, error) {
	user := new(model.User)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	user.UserID, user.UserName, user.Password, user.Phone, user.Email = row.UserID, row.UserName, row.Password, row.Phone, row.Email
	user.Nickname, user.Avatar, user.Bio, user.Gender, user.Birthday = row.Nickname, row.Avatar, row.Bio, int32(row.Gender), row.Birthday
	user.Region, user.Status, user.RegisterTime, user.LastLoginTime, user.UpdatedTime = row.Region, int32(row.Status), row.RegisterTime, row.LastLoginTime, row.UpdatedTime
	return user, nil
}
