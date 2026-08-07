package router

import (
	"net/http"

	"myim/app/service"

	"github.com/gin-gonic/gin"
)

var serv *service.Service

func New(s *service.Service) *gin.Engine {
	serv = s
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.POST("/user-register", HandleCGUserRegister)
	router.POST("/user-login", HandleCGUserLogin)
	return router
}
