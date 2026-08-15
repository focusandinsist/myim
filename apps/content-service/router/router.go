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

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) }) // 健康检查
	router.POST("/content/create", HandleCGContentCreate)                          // 创建草稿或直接发布动态
	router.POST("/content/publish", HandleCGContentPublish)                        // 发布当前用户的草稿
	router.POST("/content/get", HandleCGContentGet)                                // 查询动态详情
	router.POST("/content/list", HandleCGContentList)                              // 查询用户动态列表
	router.POST("/content/delete", HandleCGContentDelete)                          // 删除当前用户的动态
	router.POST("/content/like", HandleCGContentLike)                              // 点赞动态
	router.POST("/content/unlike", HandleCGContentUnlike)                          // 取消点赞
	router.POST("/content/comment", HandleCGContentComment)                        // 创建评论
	router.POST("/content/comments", HandleCGContentComments)                      // 查询评论列表
	router.POST("/content/comment/delete", HandleCGContentCommentDelete)           // 删除自己的评论
	router.POST("/social/follow", HandleCGSocialFollow)                            // 关注用户
	router.POST("/social/unfollow", HandleCGSocialUnfollow)                        // 取消关注用户
	return router
}
