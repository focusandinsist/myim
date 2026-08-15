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
	config    *config.Config
	dao       *contentdao.Dao
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
	return &Service{config: c, dao: d, publisher: publisher}
}
func (s *Service) user(token string) (string, error) {
	parts := strings.Fields(token)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", userservice.ErrInvalidAccessToken
	}
	claims, e := userservice.ValidateAccessToken(parts[1], s.config.JWTSecret)
	if e != nil {
		return "", e
	}
	return claims.UserID, nil
}
func toItem(c *model.Content) *proto.ContentItem {
	p := &proto.ContentItem{ContentId: c.ContentID, AuthorUserId: c.AuthorUserID, Text: c.Text, MediaUrls: c.MediaURLs, Status: c.Status, LikeCount: c.LikeCount, CommentCount: c.CommentCount, CreatedAt: c.CreatedAt.Unix(), UpdatedAt: c.UpdatedAt.Unix()}
	if c.PublishedAt != nil {
		p.PublishedAt = c.PublishedAt.Unix()
	}
	return p
}
func (s *Service) HandleCGContentCreate(ctx context.Context, in *proto.CGContentCreate, token string) (*proto.GCContentCreate, error) {
	o := new(proto.GCContentCreate)
	uid, e := s.user(token)
	if e != nil {
		o.ErrorCode = http.StatusUnauthorized
		return o, e
	}
	if in == nil || strings.TrimSpace(in.Text) == "" {
		o.ErrorCode = http.StatusBadRequest
		return o, fmt.Errorf("%w: text is required", ErrInvalidInput)
	}
	now := time.Now()
	c := &model.Content{ContentID: uuid.NewString(), AuthorUserID: uid, Text: strings.TrimSpace(in.Text), MediaURLs: in.MediaUrls, Status: model.StatusDraft, CreatedAt: now, UpdatedAt: now}
	if in.Publish {
		c.Status = model.StatusPublished
		c.PublishedAt = &now
	}
	if e = s.dao.CreateContent(ctx, c); e != nil {
		o.ErrorCode = 500
		return o, e
	}
	o.Content = toItem(c)
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGContentPublish(ctx context.Context, in *proto.CGContentPublish, token string) (*proto.GCContentPublish, error) {
	o := new(proto.GCContentPublish)
	uid, e := s.user(token)
	if e != nil {
		o.ErrorCode = 401
		return o, e
	}
	c, e := s.dao.GetContent(ctx, in.GetContentId())
	if e != nil {
		o.ErrorCode = 404
		return o, e
	}
	if c.AuthorUserID != uid {
		o.ErrorCode = 403
		return o, ErrForbidden
	}
	c, e = s.dao.PublishContent(ctx, c.ContentID, time.Now())
	if e != nil {
		o.ErrorCode = 404
		return o, e
	}
	o.Content = toItem(c)
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGContentGet(ctx context.Context, in *proto.CGContentGet, token string) (*proto.GCContentGet, error) {
	o := new(proto.GCContentGet)
	uid, e := s.user(token)
	if e != nil {
		o.ErrorCode = 401
		return o, e
	}
	c, e := s.dao.GetContent(ctx, in.GetContentId())
	if e != nil {
		o.ErrorCode = 404
		return o, e
	}
	if c.Status == model.StatusDeleted {
		o.ErrorCode = 404
		return o, contentdao.ErrNotFound
	}
	item := toItem(c)
	o.Content = item
	_ = uid
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGContentList(ctx context.Context, in *proto.CGContentList, token string) (*proto.GCContentList, error) {
	o := new(proto.GCContentList)
	if _, e := s.user(token); e != nil {
		o.ErrorCode = 401
		return o, e
	}
	if in.GetUserId() == "" {
		o.ErrorCode = 400
		return o, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	cs, e := s.dao.ListContents(ctx, in.UserId, int(in.Page), int(in.PageSize))
	if e != nil {
		o.ErrorCode = 500
		return o, e
	}
	for _, c := range cs {
		o.Contents = append(o.Contents, toItem(c))
	}
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGContentDelete(ctx context.Context, in *proto.CGContentDelete, token string) (*proto.GCContentDelete, error) {
	o := new(proto.GCContentDelete)
	uid, e := s.user(token)
	if e != nil {
		o.ErrorCode = 401
		return o, e
	}
	c, e := s.dao.GetContent(ctx, in.GetContentId())
	if e != nil {
		o.ErrorCode = 404
		return o, e
	}
	if c.AuthorUserID != uid {
		o.ErrorCode = 403
		return o, ErrForbidden
	}
	e = s.dao.DeleteContent(ctx, c.ContentID)
	if e != nil {
		o.ErrorCode = 404
	}
	o.ErrorMsg = "ok"
	return o, e
}
func (s *Service) like(ctx context.Context, id, token string, active bool) (int64, error) {
	uid, e := s.user(token)
	if e != nil {
		return 0, e
	}
	if _, e = s.dao.GetContent(ctx, id); e != nil {
		return 0, e
	}
	return s.dao.ToggleLike(ctx, id, uid, active)
}
func (s *Service) HandleCGContentLike(ctx context.Context, in *proto.CGContentLike, t string) (*proto.GCContentLike, error) {
	o := new(proto.GCContentLike)
	n, e := s.like(ctx, in.GetContentId(), t, true)
	if e != nil {
		o.ErrorCode = 401
		return o, e
	}
	o.LikeCount = n
	o.Liked = true
	uid, _ := s.user(t)
	if payload, eventErr := event.LikeEnvelope("content.liked.v1", in.GetContentId(), uid, true); eventErr == nil {
		_ = s.publisher.Publish(event.TopicContentEvents, in.GetContentId(), payload)
	}
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGContentUnlike(ctx context.Context, in *proto.CGContentUnlike, t string) (*proto.GCContentUnlike, error) {
	o := new(proto.GCContentUnlike)
	n, e := s.like(ctx, in.GetContentId(), t, false)
	if e != nil {
		o.ErrorCode = 401
		return o, e
	}
	o.LikeCount = n
	uid, _ := s.user(t)
	if payload, eventErr := event.LikeEnvelope("content.unliked.v1", in.GetContentId(), uid, false); eventErr == nil {
		_ = s.publisher.Publish(event.TopicContentEvents, in.GetContentId(), payload)
	}
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGContentComment(ctx context.Context, in *proto.CGContentComment, t string) (*proto.GCContentComment, error) {
	o := new(proto.GCContentComment)
	uid, e := s.user(t)
	if e != nil {
		o.ErrorCode = 401
		return o, e
	}
	if strings.TrimSpace(in.Text) == "" {
		o.ErrorCode = 400
		return o, ErrInvalidInput
	}
	if _, e = s.dao.GetContent(ctx, in.ContentId); e != nil {
		o.ErrorCode = 404
		return o, e
	}
	c := &model.Comment{CommentID: uuid.NewString(), ContentID: in.ContentId, AuthorUserID: uid, ParentID: in.ParentId, Text: strings.TrimSpace(in.Text), Status: model.CommentActive}
	if e = s.dao.CreateComment(ctx, c); e != nil {
		o.ErrorCode = 500
		return o, e
	}
	o.Comment = &proto.CommentItem{CommentId: c.CommentID, ContentId: c.ContentID, AuthorUserId: c.AuthorUserID, ParentId: c.ParentID, Text: c.Text, Status: c.Status, CreatedAt: c.CreatedAt.Unix()}
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGContentComments(ctx context.Context, in *proto.CGContentComments, t string) (*proto.GCContentComments, error) {
	o := new(proto.GCContentComments)
	if _, e := s.user(t); e != nil {
		o.ErrorCode = 401
		return o, e
	}
	cs, e := s.dao.ListComments(ctx, in.ContentId, int(in.Page), int(in.PageSize))
	if e != nil {
		o.ErrorCode = 500
		return o, e
	}
	for _, c := range cs {
		o.Comments = append(o.Comments, &proto.CommentItem{CommentId: c.CommentID, ContentId: c.ContentID, AuthorUserId: c.AuthorUserID, ParentId: c.ParentID, Text: c.Text, Status: c.Status, CreatedAt: c.CreatedAt.Unix()})
	}
	o.ErrorMsg = "ok"
	return o, nil
}

func (s *Service) HandleCGContentCommentDelete(ctx context.Context, in *proto.CGContentCommentDelete, t string) (*proto.GCContentCommentDelete, error) {
	o := new(proto.GCContentCommentDelete)
	uid, err := s.user(t)
	if err != nil {
		o.ErrorCode = http.StatusUnauthorized
		return o, err
	}
	if in == nil || in.CommentId == "" {
		o.ErrorCode = http.StatusBadRequest
		return o, ErrInvalidInput
	}
	if err = s.dao.DeleteComment(ctx, in.CommentId, uid); err != nil {
		o.ErrorCode = http.StatusNotFound
		return o, err
	}
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) follow(ctx context.Context, in *proto.CGSocialFollow, t string, active bool) error {
	uid, e := s.user(t)
	if e != nil {
		return e
	}
	if uid == in.UserId {
		return fmt.Errorf("%w: cannot follow self", ErrInvalidInput)
	}
	if uuid.Validate(in.UserId) != nil {
		return fmt.Errorf("%w: valid user_id is required", ErrInvalidInput)
	}
	return s.dao.Follow(ctx, uid, in.UserId, active)
}
func (s *Service) HandleCGSocialFollow(ctx context.Context, in *proto.CGSocialFollow, t string) (*proto.GCSocialFollow, error) {
	o := new(proto.GCSocialFollow)
	e := s.follow(ctx, in, t, true)
	if e != nil {
		o.ErrorCode = 400
		return o, e
	}
	o.Following = true
	o.ErrorMsg = "ok"
	return o, nil
}
func (s *Service) HandleCGSocialUnfollow(ctx context.Context, in *proto.CGSocialUnfollow, t string) (*proto.GCSocialUnfollow, error) {
	o := new(proto.GCSocialUnfollow)
	e := s.follow(ctx, &proto.CGSocialFollow{UserId: in.UserId}, t, false)
	if e != nil {
		o.ErrorCode = 400
		return o, e
	}
	o.ErrorMsg = "ok"
	return o, nil
}
