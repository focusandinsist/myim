package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	proto_social "myim/api/protobuf/social"
	"myim/apps/social-service/model"
	user_service "myim/apps/user-service/service"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid input")  // 请求字段缺失或格式不符合业务要求
	ErrUserNotFound = errors.New("user not found") // 目标用户不存在或不可用
)

const (
	defaultPage     int32 = 1   // 默认页码
	defaultPageSize int32 = 20  // 默认每页数量
	maxPageSize     int32 = 100 // 最大每页数量
)

func (s *Service) HandleCGSocialFollow(ctx context.Context, input *proto_social.CGSocialFollow, accessToken string) (output *proto_social.GCSocialFollow, err error) {
	output = new(proto_social.GCSocialFollow)
	claims, err := user_service.ValidateAccessToken(accessToken, s.config.JWTSecret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = err.Error()
		return output, err
	}
	if input == nil || uuid.Validate(input.GetTargetUserId()) != nil {
		err = fmt.Errorf("%w: valid target_user_id is required", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}
	if claims.UserID == input.GetTargetUserId() {
		err = fmt.Errorf("%w: cannot follow self", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}

	exists, err := s.dao.UserExists(ctx, input.GetTargetUserId())
	if err != nil {
		err = fmt.Errorf("check target user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	if !exists {
		output.ErrorCode = http.StatusNotFound
		output.ErrorMsg = ErrUserNotFound.Error()
		return output, ErrUserNotFound
	}

	follow, err := s.dao.SetFollow(ctx, claims.UserID, input.GetTargetUserId(), model.FollowStatusActive)
	if err != nil {
		err = fmt.Errorf("follow user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	output.ErrorMsg = "ok"
	output.Relation = toProtoFollow(follow)
	return output, nil
}

func (s *Service) HandleCGSocialUnfollow(ctx context.Context, input *proto_social.CGSocialUnfollow, accessToken string) (output *proto_social.GCSocialUnfollow, err error) {
	output = new(proto_social.GCSocialUnfollow)
	claims, err := user_service.ValidateAccessToken(accessToken, s.config.JWTSecret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = err.Error()
		return output, err
	}
	if input == nil || uuid.Validate(input.GetTargetUserId()) != nil {
		err = fmt.Errorf("%w: valid target_user_id is required", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}
	if claims.UserID == input.GetTargetUserId() {
		err = fmt.Errorf("%w: cannot unfollow self", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}
	exists, err := s.dao.UserExists(ctx, input.GetTargetUserId())
	if err != nil {
		err = fmt.Errorf("check target user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	if !exists {
		output.ErrorCode = http.StatusNotFound
		output.ErrorMsg = ErrUserNotFound.Error()
		return output, ErrUserNotFound
	}

	follow, err := s.dao.SetFollow(ctx, claims.UserID, input.GetTargetUserId(), model.FollowStatusCancelled)
	if err != nil {
		err = fmt.Errorf("unfollow user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	output.ErrorMsg = "ok"
	output.Relation = toProtoFollow(follow)
	return output, nil
}

func (s *Service) HandleCGSocialFollowingList(ctx context.Context, input *proto_social.CGSocialFollowingList, accessToken string) (output *proto_social.GCSocialFollowingList, err error) {
	output = new(proto_social.GCSocialFollowingList)
	claims, err := user_service.ValidateAccessToken(accessToken, s.config.JWTSecret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = err.Error()
		return output, err
	}
	page, pageSize := normalizePage(input.GetPage(), input.GetPageSize())
	follows, err := s.dao.ListFollowing(ctx, claims.UserID, page, pageSize)
	if err != nil {
		err = fmt.Errorf("list following: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	for _, follow := range follows {
		output.Relations = append(output.Relations, toProtoFollow(follow))
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGSocialFollowerList(ctx context.Context, input *proto_social.CGSocialFollowerList, accessToken string) (output *proto_social.GCSocialFollowerList, err error) {
	output = new(proto_social.GCSocialFollowerList)
	claims, err := user_service.ValidateAccessToken(accessToken, s.config.JWTSecret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = err.Error()
		return output, err
	}
	page, pageSize := normalizePage(input.GetPage(), input.GetPageSize())
	follows, err := s.dao.ListFollowers(ctx, claims.UserID, page, pageSize)
	if err != nil {
		err = fmt.Errorf("list followers: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	for _, follow := range follows {
		output.Relations = append(output.Relations, toProtoFollow(follow))
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGSocialFollowCheck(ctx context.Context, input *proto_social.CGSocialFollowCheck, accessToken string) (output *proto_social.GCSocialFollowCheck, err error) {
	output = new(proto_social.GCSocialFollowCheck)
	claims, err := user_service.ValidateAccessToken(accessToken, s.config.JWTSecret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = err.Error()
		return output, err
	}
	if input == nil || uuid.Validate(input.GetTargetUserId()) != nil {
		err = fmt.Errorf("%w: valid target_user_id is required", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}
	exists, err := s.dao.UserExists(ctx, input.GetTargetUserId())
	if err != nil {
		err = fmt.Errorf("check target user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	if !exists {
		output.ErrorCode = http.StatusNotFound
		output.ErrorMsg = ErrUserNotFound.Error()
		return output, ErrUserNotFound
	}
	following, err := s.dao.IsFollowing(ctx, claims.UserID, input.GetTargetUserId())
	if err != nil {
		err = fmt.Errorf("check following: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	output.ErrorMsg = "ok"
	output.Following = following
	return output, nil
}

func normalizePage(page, pageSize int32) (int32, int32) {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}
	return page, pageSize
}

func toProtoFollow(follow *model.Follow) *proto_social.FollowRelation {
	return &proto_social.FollowRelation{
		FollowerUserId: follow.FollowerUserID,
		FolloweeUserId: follow.FolloweeUserID,
		Status:         proto_social.FollowStatus(follow.Status),
		CreatedAt:      follow.CreatedAt.Unix(),
		UpdatedAt:      follow.UpdatedAt.Unix(),
	}
}
