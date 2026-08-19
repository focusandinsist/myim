package dao

import (
	"context"
	"fmt"

	"myim/apps/social-service/model"
)

func (d *Dao) UserExists(ctx context.Context, userID string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE user_id = $1 AND status = 0
		)
	`
	var exists bool
	if err := d.db.QueryRowContext(ctx, query, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists, nil
}

func (d *Dao) SetFollow(ctx context.Context, followerUserID, followeeUserID string, status int32) (*model.Follow, error) {
	const query = `
		INSERT INTO follows
			(follower_user_id, followee_user_id, status)
		VALUES
			($1, $2, $3)
		ON CONFLICT (follower_user_id, followee_user_id)
		DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = CURRENT_TIMESTAMP
		RETURNING
			follower_user_id, followee_user_id, status, created_at, updated_at
	`
	follow := new(model.Follow)
	if err := d.db.QueryRowContext(ctx, query, followerUserID, followeeUserID, status).Scan(
		&follow.FollowerUserID,
		&follow.FolloweeUserID,
		&follow.Status,
		&follow.CreatedAt,
		&follow.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("set follow relation: %w", err)
	}
	return follow, nil
}

func (d *Dao) ListFollowing(ctx context.Context, followerUserID string, page, pageSize int32) ([]*model.Follow, error) {
	const query = `
		SELECT
			follower_user_id, followee_user_id, status, created_at, updated_at
		FROM follows
		WHERE follower_user_id = $1 AND status = 1
		ORDER BY updated_at DESC, followee_user_id ASC
		LIMIT $2 OFFSET $3
	`
	return d.listFollows(ctx, query, followerUserID, page, pageSize)
}

func (d *Dao) ListFollowers(ctx context.Context, followeeUserID string, page, pageSize int32) ([]*model.Follow, error) {
	const query = `
		SELECT
			follower_user_id, followee_user_id, status, created_at, updated_at
		FROM follows
		WHERE followee_user_id = $1 AND status = 1
		ORDER BY updated_at DESC, follower_user_id ASC
		LIMIT $2 OFFSET $3
	`
	return d.listFollows(ctx, query, followeeUserID, page, pageSize)
}

func (d *Dao) listFollows(ctx context.Context, query, userID string, page, pageSize int32) ([]*model.Follow, error) {
	rows, err := d.db.QueryContext(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list follow relations: %w", err)
	}
	defer rows.Close()

	var follows []*model.Follow
	for rows.Next() {
		follow := new(model.Follow)
		if err := rows.Scan(
			&follow.FollowerUserID,
			&follow.FolloweeUserID,
			&follow.Status,
			&follow.CreatedAt,
			&follow.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan follow relation: %w", err)
		}
		follows = append(follows, follow)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate follow relations: %w", err)
	}
	return follows, nil
}

func (d *Dao) IsFollowing(ctx context.Context, followerUserID, followeeUserID string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM follows
			WHERE follower_user_id = $1
				AND followee_user_id = $2
				AND status = 1
		)
	`
	var following bool
	if err := d.db.QueryRowContext(ctx, query, followerUserID, followeeUserID).Scan(&following); err != nil {
		return false, fmt.Errorf("check follow relation: %w", err)
	}
	return following, nil
}
