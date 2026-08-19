package router

import (
	"net/http"
	"strings"

	proto_social "myim/api/protobuf/social"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

func HandleCGSocialFollow(c *gin.Context) {
	input := new(proto_social.CGSocialFollow)
	output := new(proto_social.GCSocialFollow)
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
	input := new(proto_social.CGSocialUnfollow)
	output := new(proto_social.GCSocialUnfollow)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGSocialUnfollow(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGSocialFollowingList(c *gin.Context) {
	input := new(proto_social.CGSocialFollowingList)
	output := new(proto_social.GCSocialFollowingList)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGSocialFollowingList(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGSocialFollowerList(c *gin.Context) {
	input := new(proto_social.CGSocialFollowerList)
	output := new(proto_social.GCSocialFollowerList)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGSocialFollowerList(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func HandleCGSocialFollowCheck(c *gin.Context) {
	input := new(proto_social.CGSocialFollowCheck)
	output := new(proto_social.GCSocialFollowCheck)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGSocialFollowCheck(c.Request.Context(), input, accessToken(c))
	httpx.Response(c, output, err)
}

func accessToken(c *gin.Context) string {
	authorization := strings.Fields(c.GetHeader("Authorization"))
	if len(authorization) == 2 && strings.EqualFold(authorization[0], "Bearer") {
		return authorization[1]
	}
	return ""
}
