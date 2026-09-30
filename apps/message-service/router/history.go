package router

import (
	"io"
	"net/http"
	"strings"

	proto_message "myim/api/protobuf/message"
	"myim/internal/auth"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func HandleCGConversationList(c *gin.Context) {
	input := new(proto_message.CGConversationList)
	if err := bindMessageRequest(c, input); err != nil {
		output := &proto_message.GCConversationList{ErrorCode: http.StatusBadRequest}
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := messageService.HandleCGConversationList(c.Request.Context(), input, auth.BearerToken(c.GetHeader("Authorization")))
	httpx.Response(c, output, err)
}

func HandleCGMessageHistory(c *gin.Context) {
	input := new(proto_message.CGMessageHistory)
	if err := bindMessageRequest(c, input); err != nil {
		output := &proto_message.GCMessageHistory{ErrorCode: http.StatusBadRequest}
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := messageService.HandleCGMessageHistory(c.Request.Context(), input, auth.BearerToken(c.GetHeader("Authorization")))
	httpx.Response(c, output, err)
}

func bindMessageRequest(c *gin.Context, input proto.Message) error {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 8193))
	if err != nil {
		return err
	}
	if len(body) == 0 || len(body) > 8192 {
		return httpx.ErrInvalidRequest
	}
	contentType := strings.ToLower(c.ContentType())
	if contentType == "application/x-protobuf" || contentType == "application/protobuf" {
		return proto.Unmarshal(body, input)
	}
	if contentType != "application/json" {
		return httpx.ErrInvalidRequest
	}
	return protojson.Unmarshal(body, input)
}
