package router

import (
	"context"
	"net/http"

	"myim/apps/content-service/service"
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
	router.POST("/content/create", HandleCGContentCreate)                                                   // 创建草稿或直接发布动态
	router.POST("/content/publish", HandleCGContentPublish)                                                 // 发布当前用户的草稿
	router.POST("/content/get", HandleCGContentGet)                                                         // 查询动态详情
	router.POST("/content/list", HandleCGContentList)                                                       // 查询用户动态列表
	router.POST("/content/delete", HandleCGContentDelete)                                                   // 删除当前用户的动态
	router.POST("/content/like", HandleCGContentLike)                                                       // 点赞动态
	router.POST("/content/unlike", HandleCGContentUnlike)                                                   // 取消点赞
	router.POST("/content/comment", HandleCGContentComment)                                                 // 创建评论
	router.POST("/content/comments", HandleCGContentComments)                                               // 查询评论列表
	router.POST("/content/comment/delete", HandleCGContentCommentDelete)                                    // 删除自己的评论
	return router
}

func Wait(ctx context.Context) error {
	if metrics == nil {
		return nil
	}
	return metrics.Wait(ctx)
}
