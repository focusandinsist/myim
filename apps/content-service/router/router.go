package router

import (
	"net/http"
	"strings"

	proto "myim/api/protobuf/content"
	"myim/apps/content-service/service"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

var serv *service.Service

func New(s *service.Service) *gin.Engine {
	serv = s
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.POST("/content/create", handle(func() any { return new(proto.CGContentCreate) }, func(c *gin.Context, in any, token string) (any, error) {
		return serv.HandleCGContentCreate(c, in.(*proto.CGContentCreate), token)
	}))
	r.POST("/content/publish", handle(func() any { return new(proto.CGContentPublish) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentPublish(c, in.(*proto.CGContentPublish), t)
	}))
	r.POST("/content/get", handle(func() any { return new(proto.CGContentGet) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentGet(c, in.(*proto.CGContentGet), t)
	}))
	r.POST("/content/list", handle(func() any { return new(proto.CGContentList) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentList(c, in.(*proto.CGContentList), t)
	}))
	r.POST("/content/delete", handle(func() any { return new(proto.CGContentDelete) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentDelete(c, in.(*proto.CGContentDelete), t)
	}))
	r.POST("/content/like", handle(func() any { return new(proto.CGContentLike) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentLike(c, in.(*proto.CGContentLike), t)
	}))
	r.POST("/content/unlike", handle(func() any { return new(proto.CGContentUnlike) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentUnlike(c, in.(*proto.CGContentUnlike), t)
	}))
	r.POST("/content/comment", handle(func() any { return new(proto.CGContentComment) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentComment(c, in.(*proto.CGContentComment), t)
	}))
	r.POST("/content/comments", handle(func() any { return new(proto.CGContentComments) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentComments(c, in.(*proto.CGContentComments), t)
	}))
	r.POST("/content/comment/delete", handle(func() any { return new(proto.CGContentCommentDelete) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGContentCommentDelete(c, in.(*proto.CGContentCommentDelete), t)
	}))
	r.POST("/social/follow", handle(func() any { return new(proto.CGSocialFollow) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGSocialFollow(c, in.(*proto.CGSocialFollow), t)
	}))
	r.POST("/social/unfollow", handle(func() any { return new(proto.CGSocialUnfollow) }, func(c *gin.Context, in any, t string) (any, error) {
		return serv.HandleCGSocialUnfollow(c, in.(*proto.CGSocialUnfollow), t)
	}))
	return r
}

type runner func(*gin.Context, any, string) (any, error)

func handle(factory func() any, run runner) gin.HandlerFunc {
	return func(c *gin.Context) {
		in := factory()
		if err := c.ShouldBind(in); err != nil {
			httpx.Response(c, in, err)
			return
		}
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		out, e := run(c, in, token)
		httpx.Response(c, out, e)
	}
}
