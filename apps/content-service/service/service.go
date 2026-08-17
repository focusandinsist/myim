package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	proto "myim/api/protobuf/content"
	"myim/apps/content-service/config"
	contentdao "myim/apps/content-service/dao"
	"myim/apps/content-service/event"
	"myim/apps/content-service/model"
	userservice "myim/apps/user-service/service"
)

var (
	ErrInvalidInput = errors.New("invalid input")  // 请求字段缺失或格式不符合业务要求
	ErrForbidden    = errors.New("forbidden")      // 当前用户没有修改目标资源的权限
	ErrUserNotFound = errors.New("user not found") // 目标用户不存在
)

type Service struct {
	config    *config.Config  // Content服务配置
	dao       *contentdao.Dao // Content PostgreSQL数据访问对象
	publisher event.Publisher // Content事件发布器
}

func (s *Service) Stop() {
	if publisher, ok := s.publisher.(interface{ Close() error }); ok {
		_ = publisher.Close()
	}
}

func New(c *config.Config, d *contentdao.Dao) *Service {
	publisher := event.Publisher(event.LogPublisher{})
	if kafkaPublisher, err := event.NewSaramaPublisher(c.KafkaBrokers); err == nil {
		publisher = kafkaPublisher
	}
	return &Service{
		config:    c,
		dao:       d,
		publisher: publisher,
	}
}

func (s *Service) user(token string) (string, error) {
	parts := strings.Fields(token)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", userservice.ErrInvalidAccessToken
	}
	claims, err := userservice.ValidateAccessToken(parts[1], s.config.JWTSecret)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func toItem(c *model.Content) *proto.ContentItem {
	item := &proto.ContentItem{
		ContentId:    c.ContentID,
		AuthorUserId: c.AuthorUserID,
		Text:         c.Text,
		MediaUrls:    c.MediaURLs,
		Status:       c.Status,
		LikeCount:    c.LikeCount,
		CommentCount: c.CommentCount,
		CreatedAt:    c.CreatedAt.Unix(),
		UpdatedAt:    c.UpdatedAt.Unix(),
	}
	if c.PublishedAt != nil {
		item.PublishedAt = c.PublishedAt.Unix()
	}
	return item
}

func (s *Service) HandleCGContentCreate(ctx context.Context, in *proto.CGContentCreate, token string) (output *proto.GCContentCreate, err error) {
	output = new(proto.GCContentCreate)
	userID, err := s.user(token)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	if in == nil || strings.TrimSpace(in.Text) == "" {
		output.ErrorCode = http.StatusBadRequest
		return output, fmt.Errorf("%w: text is required", ErrInvalidInput)
	}
	now := time.Now()
	content := &model.Content{
		ContentID:    uuid.NewString(),
		AuthorUserID: userID,
		Text:         strings.TrimSpace(in.Text),
		MediaURLs:    in.MediaUrls,
		Status:       model.StatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if in.Publish {
		content.Status = model.StatusPublished
		content.PublishedAt = &now
	}
	if err = s.dao.CreateContent(ctx, content); err != nil {
		output.ErrorCode = http.StatusInternalServerError
		return output, err
	}
	output.Content = toItem(content)
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentPublish(ctx context.Context, in *proto.CGContentPublish, token string) (output *proto.GCContentPublish, err error) {
	output = new(proto.GCContentPublish)
	userID, err := s.user(token)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	content, err := s.dao.GetContent(ctx, in.GetContentId())
	if err != nil {
		output.ErrorCode = http.StatusNotFound
		return output, err
	}
	if content.AuthorUserID != userID {
		output.ErrorCode = http.StatusForbidden
		return output, ErrForbidden
	}
	content, err = s.dao.PublishContent(ctx, content.ContentID, time.Now())
	if err != nil {
		output.ErrorCode = http.StatusNotFound
		return output, err
	}
	output.Content = toItem(content)
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentGet(ctx context.Context, in *proto.CGContentGet, token string) (output *proto.GCContentGet, err error) {
	output = new(proto.GCContentGet)
	_, err = s.user(token)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	content, err := s.dao.GetContent(ctx, in.GetContentId())
	if err != nil {
		output.ErrorCode = http.StatusNotFound
		return output, err
	}
	if content.Status == model.StatusDeleted {
		output.ErrorCode = http.StatusNotFound
		return output, contentdao.ErrNotFound
	}
	output.Content = toItem(content)
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentList(ctx context.Context, in *proto.CGContentList, token string) (output *proto.GCContentList, err error) {
	output = new(proto.GCContentList)
	if _, err = s.user(token); err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	if in.GetUserId() == "" {
		output.ErrorCode = http.StatusBadRequest
		return output, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	contents, err := s.dao.ListContents(ctx, in.UserId, int(in.Page), int(in.PageSize))
	if err != nil {
		output.ErrorCode = http.StatusInternalServerError
		return output, err
	}
	for _, content := range contents {
		output.Contents = append(output.Contents, toItem(content))
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentDelete(ctx context.Context, in *proto.CGContentDelete, token string) (output *proto.GCContentDelete, err error) {
	output = new(proto.GCContentDelete)
	userID, err := s.user(token)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	content, err := s.dao.GetContent(ctx, in.GetContentId())
	if err != nil {
		output.ErrorCode = http.StatusNotFound
		return output, err
	}
	if content.AuthorUserID != userID {
		output.ErrorCode = http.StatusForbidden
		return output, ErrForbidden
	}
	err = s.dao.DeleteContent(ctx, content.ContentID)
	if err != nil {
		output.ErrorCode = http.StatusNotFound
	}
	output.ErrorMsg = "ok"
	return output, err
}

func (s *Service) like(ctx context.Context, id, token string, active bool) (int64, error) {
	userID, err := s.user(token)
	if err != nil {
		return 0, err
	}
	if _, err = s.dao.GetContent(ctx, id); err != nil {
		return 0, err
	}
	return s.dao.ToggleLike(ctx, id, userID, active)
}

func (s *Service) HandleCGContentLike(ctx context.Context, in *proto.CGContentLike, accessToken string) (output *proto.GCContentLike, err error) {
	output = new(proto.GCContentLike)
	likeCount, err := s.like(ctx, in.GetContentId(), accessToken, true)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	output.LikeCount = likeCount
	output.Liked = true
	userID, _ := s.user(accessToken)
	if payload, eventErr := event.LikeEnvelope("content.liked.v1", in.GetContentId(), userID, true); eventErr == nil {
		_ = s.publisher.Publish(event.TopicContentEvents, in.GetContentId(), payload)
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentUnlike(ctx context.Context, in *proto.CGContentUnlike, accessToken string) (output *proto.GCContentUnlike, err error) {
	output = new(proto.GCContentUnlike)
	likeCount, err := s.like(ctx, in.GetContentId(), accessToken, false)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	output.LikeCount = likeCount
	userID, _ := s.user(accessToken)
	if payload, eventErr := event.LikeEnvelope("content.unliked.v1", in.GetContentId(), userID, false); eventErr == nil {
		_ = s.publisher.Publish(event.TopicContentEvents, in.GetContentId(), payload)
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentComment(ctx context.Context, in *proto.CGContentComment, accessToken string) (output *proto.GCContentComment, err error) {
	output = new(proto.GCContentComment)
	userID, err := s.user(accessToken)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	if strings.TrimSpace(in.Text) == "" {
		output.ErrorCode = http.StatusBadRequest
		return output, ErrInvalidInput
	}
	if _, err = s.dao.GetContent(ctx, in.ContentId); err != nil {
		output.ErrorCode = http.StatusNotFound
		return output, err
	}
	comment := &model.Comment{
		CommentID:    uuid.NewString(),
		ContentID:    in.ContentId,
		AuthorUserID: userID,
		ParentID:     in.ParentId,
		Text:         strings.TrimSpace(in.Text),
		Status:       model.CommentActive,
	}
	if err = s.dao.CreateComment(ctx, comment); err != nil {
		output.ErrorCode = http.StatusInternalServerError
		return output, err
	}
	output.Comment = &proto.CommentItem{
		CommentId:    comment.CommentID,
		ContentId:    comment.ContentID,
		AuthorUserId: comment.AuthorUserID,
		ParentId:     comment.ParentID,
		Text:         comment.Text,
		Status:       comment.Status,
		CreatedAt:    comment.CreatedAt.Unix(),
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentComments(ctx context.Context, in *proto.CGContentComments, accessToken string) (output *proto.GCContentComments, err error) {
	output = new(proto.GCContentComments)
	if _, err = s.user(accessToken); err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	comments, err := s.dao.ListComments(ctx, in.ContentId, int(in.Page), int(in.PageSize))
	if err != nil {
		output.ErrorCode = http.StatusInternalServerError
		return output, err
	}
	for _, comment := range comments {
		output.Comments = append(output.Comments, &proto.CommentItem{
			CommentId:    comment.CommentID,
			ContentId:    comment.ContentID,
			AuthorUserId: comment.AuthorUserID,
			ParentId:     comment.ParentID,
			Text:         comment.Text,
			Status:       comment.Status,
			CreatedAt:    comment.CreatedAt.Unix(),
		})
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGContentCommentDelete(ctx context.Context, in *proto.CGContentCommentDelete, accessToken string) (output *proto.GCContentCommentDelete, err error) {
	output = new(proto.GCContentCommentDelete)
	userID, err := s.user(accessToken)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	if in == nil || in.CommentId == "" {
		output.ErrorCode = http.StatusBadRequest
		return output, ErrInvalidInput
	}
	if err = s.dao.DeleteComment(ctx, in.CommentId, userID); err != nil {
		output.ErrorCode = http.StatusNotFound
		return output, err
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) follow(ctx context.Context, in *proto.CGSocialFollow, accessToken string, active bool) error {
	userID, err := s.user(accessToken)
	if err != nil {
		return err
	}
	if userID == in.UserId {
		return fmt.Errorf("%w: cannot follow self", ErrInvalidInput)
	}
	if uuid.Validate(in.UserId) != nil {
		return fmt.Errorf("%w: valid user_id is required", ErrInvalidInput)
	}
	return s.dao.Follow(ctx, userID, in.UserId, active)
}

func (s *Service) HandleCGSocialFollow(ctx context.Context, in *proto.CGSocialFollow, accessToken string) (output *proto.GCSocialFollow, err error) {
	output = new(proto.GCSocialFollow)
	err = s.follow(ctx, in, accessToken, true)
	if err != nil {
		output.ErrorCode = http.StatusBadRequest
		return output, err
	}
	output.Following = true
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGSocialUnfollow(ctx context.Context, in *proto.CGSocialUnfollow, accessToken string) (output *proto.GCSocialUnfollow, err error) {
	output = new(proto.GCSocialUnfollow)
	err = s.follow(ctx, &proto.CGSocialFollow{UserId: in.UserId}, accessToken, false)
	if err != nil {
		output.ErrorCode = http.StatusBadRequest
		return output, err
	}
	output.ErrorMsg = "ok"
	return output, nil
}
