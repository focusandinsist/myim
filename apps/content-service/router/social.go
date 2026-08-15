package router

import (
	"net/http"

	proto "myim/api/protobuf/content"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

func HandleCGSocialFollow(c *gin.Context) {
	input := new(proto.CGSocialFollow)
	output := new(proto.GCSocialFollow)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGSocialFollow(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGSocialUnfollow(c *gin.Context) {
	input := new(proto.CGSocialUnfollow)
	output := new(proto.GCSocialUnfollow)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGSocialUnfollow(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}
