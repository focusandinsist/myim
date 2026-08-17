package router

import (
	"net/http"
	"strings"

	proto "myim/api/protobuf/content"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

func accessToken(c *gin.Context) string {
	authorization := strings.Fields(c.GetHeader("Authorization"))
	if len(authorization) == 2 && strings.EqualFold(authorization[0], "Bearer") {
		return authorization[1]
	}
	return ""
}

func HandleCGContentCreate(c *gin.Context) {
	input := new(proto.CGContentCreate)
	output := new(proto.GCContentCreate)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentCreate(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentPublish(c *gin.Context) {
	input := new(proto.CGContentPublish)
	output := new(proto.GCContentPublish)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentPublish(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentGet(c *gin.Context) {
	input := new(proto.CGContentGet)
	output := new(proto.GCContentGet)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentGet(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentList(c *gin.Context) {
	input := new(proto.CGContentList)
	output := new(proto.GCContentList)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentList(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentDelete(c *gin.Context) {
	input := new(proto.CGContentDelete)
	output := new(proto.GCContentDelete)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentDelete(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentLike(c *gin.Context) {
	input := new(proto.CGContentLike)
	output := new(proto.GCContentLike)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentLike(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentUnlike(c *gin.Context) {
	input := new(proto.CGContentUnlike)
	output := new(proto.GCContentUnlike)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentUnlike(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentComment(c *gin.Context) {
	input := new(proto.CGContentComment)
	output := new(proto.GCContentComment)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentComment(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentComments(c *gin.Context) {
	input := new(proto.CGContentComments)
	output := new(proto.GCContentComments)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentComments(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGContentCommentDelete(c *gin.Context) {
	input := new(proto.CGContentCommentDelete)
	output := new(proto.GCContentCommentDelete)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGContentCommentDelete(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}
