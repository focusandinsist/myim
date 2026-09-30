package router

import (
	"context"
	"net/http"

	"myim/apps/social-service/service"
	"myim/internal/observability"

	"github.com/gin-gonic/gin"
)

var serv *service.Service
var metrics *observability.HTTP

func New(s *service.Service) *gin.Engine {
	serv = s
	router := gin.New()
	metrics = observability.NewHTTP()
	router.Use(metrics.Middleware(), gin.Recovery())
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })                          // 检查进程存活
	router.GET("/ready", observability.Health(func(ctx context.Context) error { return serv.Health(ctx) })) // 检查数据库连接并确认可接流量
	router.GET("/metrics", metrics.Metrics)                                                                 // 导出HTTP请求指标
	router.POST("/social/follow", HandleCGSocialFollow)                                                     // 关注用户
	router.POST("/social/unfollow", HandleCGSocialUnfollow)                                                 // 取消关注用户
	router.POST("/social/following-list", HandleCGSocialFollowingList)                                      // 查询当前用户关注列表
	router.POST("/social/follower-list", HandleCGSocialFollowerList)                                        // 查询当前用户粉丝列表
	router.POST("/social/follow-check", HandleCGSocialFollowCheck)                                          // 判断当前用户是否关注目标用户
	return router
}

func Wait(ctx context.Context) error {
	if metrics == nil {
		return nil
	}
	return metrics.Wait(ctx)
}
