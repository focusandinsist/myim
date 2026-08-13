package router

import (
	"net/http"
	"strings"

	proto_user "myim/api/protobuf/user"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

func HandleCGUserRegister(c *gin.Context) {
	input := new(proto_user.CGUserRegister)
	output := new(proto_user.GCUserRegister)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGUserRegister(c.Request.Context(), input)
	httpx.Response(c, output, err)
}

func HandleCGUserLogin(c *gin.Context) {
	input := new(proto_user.CGUserLogin)
	output := new(proto_user.GCUserLogin)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGUserLogin(c.Request.Context(), input)
	httpx.Response(c, output, err)
}

func HandleCGMyProfile(c *gin.Context) {
	accessToken := ""
	authorization := strings.Fields(c.GetHeader("Authorization"))
	if len(authorization) == 2 && strings.EqualFold(authorization[0], "Bearer") {
		accessToken = authorization[1]
	}
	output, err := serv.HandleCGMyProfile(c.Request.Context(), accessToken)
	httpx.Response(c, output, err)
}

func HandleCGTargetProfile(c *gin.Context) {
	input := new(proto_user.CGTargetProfile)
	output := new(proto_user.GCTargetProfile)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	accessToken := ""
	authorization := strings.Fields(c.GetHeader("Authorization"))
	if len(authorization) == 2 && strings.EqualFold(authorization[0], "Bearer") {
		accessToken = authorization[1]
	}
	output, err := serv.HandleCGTargetProfile(c.Request.Context(), input, accessToken)
	httpx.Response(c, output, err)
}
