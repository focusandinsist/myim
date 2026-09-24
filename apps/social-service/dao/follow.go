package dao

import (
	"context"
	"fmt"

	"myim/apps/social-service/model"
	socialdb "myim/internal/db/social"
)

func (d *Dao) UserExists(ctx context.Context, userID string) (bool, error) {
	exists, err := d.queries.UserExists(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists, nil
}

func (d *Dao) SetFollow(ctx context.Context, followerUserID, followeeUserID string, status int32) (*model.Follow, error) {
	row, err := d.queries.SetFollow(ctx, socialdb.SetFollowParams{FollowerUserID: followerUserID, FolloweeUserID: followeeUserID, Status: int16(status)})
	if err != nil {
		return nil, fmt.Errorf("set follow relation: %w", err)
	}
	return followFromRow(row), nil
}

func (d *Dao) ListFollowing(ctx context.Context, followerUserID string, page, pageSize int32) ([]*model.Follow, error) {
	rows, err := d.queries.ListFollowing(ctx, socialdb.ListFollowingParams{FollowerUserID: followerUserID, Limit: pageSize, Offset: (page - 1) * pageSize})
	if err != nil {
		return nil, fmt.Errorf("list follow relations: %w", err)
	}
	return followRows(rows), nil
}

func (d *Dao) ListFollowers(ctx context.Context, followeeUserID string, page, pageSize int32) ([]*model.Follow, error) {
	rows, err := d.queries.ListFollowers(ctx, socialdb.ListFollowersParams{FolloweeUserID: followeeUserID, Limit: pageSize, Offset: (page - 1) * pageSize})
	if err != nil {
		return nil, fmt.Errorf("list follow relations: %w", err)
	}
	return followRows(rows), nil
}

func (d *Dao) IsFollowing(ctx context.Context, followerUserID, followeeUserID string) (bool, error) {
	following, err := d.queries.IsFollowing(ctx, socialdb.IsFollowingParams{FollowerUserID: followerUserID, FolloweeUserID: followeeUserID})
	if err != nil {
		return false, fmt.Errorf("check follow relation: %w", err)
	}
	return following, nil
}

func followRows(rows []socialdb.Follow) []*model.Follow {
	result := make([]*model.Follow, 0, len(rows))
	for _, row := range rows {
		result = append(result, followFromRow(row))
	}
	return result
}

func followFromRow(row socialdb.Follow) *model.Follow {
	return &model.Follow{FollowerUserID: row.FollowerUserID, FolloweeUserID: row.FolloweeUserID, Status: int32(row.Status), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
