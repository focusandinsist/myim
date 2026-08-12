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

	return router
}
