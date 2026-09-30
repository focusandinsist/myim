package router

import (
	"net/http"
	"time"

	proto_message "myim/api/protobuf/message"
	"myim/internal/auth"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func HandleCGConnection(c *gin.Context) {
	output := new(proto_message.GCWSMessage)
	accessToken, tokenErr := auth.RequireBearerToken(c.GetHeader("Authorization"))
	if tokenErr != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = auth.ErrInvalidAccessToken.Error()
		httpx.Response(c, output, auth.ErrInvalidAccessToken)
		return
	}
	claims, err := auth.Validate(accessToken, messageConfig.Auth.Secret)
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
