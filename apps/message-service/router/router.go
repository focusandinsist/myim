package router

import (
	"net/http"

	"myim/apps/message-service/config"
	"myim/apps/message-service/service"

	"github.com/gin-gonic/gin"
)

var messageConfig *config.Config
var messageService *service.Service

func New(conf *config.Config, serv *service.Service) *gin.Engine {
	messageConfig = conf
	messageService = serv
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	router.GET("/ws", HandleCGConnection)
	router.POST("/message/conversations", HandleCGConversationList) // 列出当前用户的单聊会话
	router.POST("/message/history", HandleCGMessageHistory)         // 按会话序号分页读取历史消息

	return router
}
