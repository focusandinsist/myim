package router

import (
	"net/http"

	"myim/apps/social-service/service"

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
	router.POST("/social/follow", HandleCGSocialFollow)                            // 关注用户
	router.POST("/social/unfollow", HandleCGSocialUnfollow)                        // 取消关注用户
	router.POST("/social/following-list", HandleCGSocialFollowingList)             // 查询当前用户关注列表
	router.POST("/social/follower-list", HandleCGSocialFollowerList)               // 查询当前用户粉丝列表
	router.POST("/social/follow-check", HandleCGSocialFollowCheck)                 // 判断当前用户是否关注目标用户
	return router
}
