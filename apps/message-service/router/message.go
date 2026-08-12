package router

import (
	"net/http"
	"strings"
	"time"

	proto_message "myim/api/protobuf/message"
	"myim/apps/message-service/service"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func HandleCGConnection(c *gin.Context) {
	output := new(proto_message.GCWSMessage)
	authorization := strings.Fields(c.GetHeader("Authorization"))
	if len(authorization) != 2 || !strings.EqualFold(authorization[0], "Bearer") {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = service.ErrInvalidAccessToken.Error()
		httpx.Response(c, output, service.ErrInvalidAccessToken)
		return
	}
	claims, err := service.ValidateAccessToken(authorization[1], messageConfig.JWTSecret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = err.Error()
		httpx.Response(c, output, err)
		return
	}
	upgrader := websocket.Upgrader{HandshakeTimeout: 5 * time.Second}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	messageService.HandleCGConnection(c.Request.Context(), claims.UserID, conn)
}
