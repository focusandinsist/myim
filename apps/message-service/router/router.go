package router

import (
	"context"
	"net/http"

	"myim/apps/message-service/config"
	"myim/apps/message-service/service"
	"myim/internal/observability"

	"github.com/gin-gonic/gin"
)

var messageConfig *config.Config
var messageService *service.Service
var metrics *observability.HTTP

func New(conf *config.Config, serv *service.Service) *gin.Engine {
	messageConfig = conf
	messageService = serv
	router := gin.New()
	metrics = observability.NewHTTP()
	router.Use(metrics.Middleware(), gin.Recovery())
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })                                    // 检查进程存活
	router.GET("/ready", observability.Health(func(ctx context.Context) error { return messageService.Health(ctx) })) // 检查数据库连接并确认可接流量
	router.GET("/metrics", metrics.Metrics)                                                                           // 导出HTTP请求指标
	router.GET("/ws", HandleCGConnection)
	router.POST("/message/conversations", HandleCGConversationList) // 列出当前用户的单聊会话
	router.POST("/message/history", HandleCGMessageHistory)         // 按会话序号分页读取历史消息

	return router
}

func Wait(ctx context.Context) error {
	if metrics == nil {
		return nil
	}
	return metrics.Wait(ctx)
}
