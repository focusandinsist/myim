package router

import (
	"net/http"
	"strings"
	"time"

	proto_message "myim/api/protobuf/message"
	"myim/apps/message-service/config"
	"myim/apps/message-service/service"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func New(conf *config.Config, serv *service.Service) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	upgrader := websocket.Upgrader{HandshakeTimeout: 5 * time.Second}
	router.GET("/ws", func(c *gin.Context) {
		output := new(proto_message.GCWSMessage)
		authorization := strings.Fields(c.GetHeader("Authorization"))
		if len(authorization) != 2 || !strings.EqualFold(authorization[0], "Bearer") {
			output.ErrorCode = http.StatusUnauthorized
			output.ErrorMsg = service.ErrInvalidAccessToken.Error()
			httpx.Response(c, output, service.ErrInvalidAccessToken)
			return
		}
		claims, err := service.ValidateAccessToken(authorization[1], conf.JWTSecret)
		if err != nil {
			output.ErrorCode = http.StatusUnauthorized
			output.ErrorMsg = err.Error()
			httpx.Response(c, output, err)
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		serv.HandleCGConnection(claims.UserID, conn)
	})

	return router
}
