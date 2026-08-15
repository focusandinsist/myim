package router

import (
	"net/http"

	"myim/apps/content-service/service"

	"github.com/gin-gonic/gin"
)

var serv *service.Service

func New(s *service.Service) *gin.Engine {
	serv = s
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	router.POST("/content/create", HandleCGContentCreate)
	router.POST("/content/publish", HandleCGContentPublish)
	router.POST("/content/get", HandleCGContentGet)
	router.POST("/content/list", HandleCGContentList)
	router.POST("/content/delete", HandleCGContentDelete)
	router.POST("/content/like", HandleCGContentLike)
	router.POST("/content/unlike", HandleCGContentUnlike)
	router.POST("/content/comment", HandleCGContentComment)
	router.POST("/content/comments", HandleCGContentComments)
	router.POST("/content/comment/delete", HandleCGContentCommentDelete)
	router.POST("/social/follow", HandleCGSocialFollow)
	router.POST("/social/unfollow", HandleCGSocialUnfollow)
	return router
}
