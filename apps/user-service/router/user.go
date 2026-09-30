package router

import (
	proto_user "myim/api/protobuf/user"
	"myim/internal/auth"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

func HandleCGUserRegister(c *gin.Context) {
	input := new(proto_user.CGUserRegister)
	output := new(proto_user.GCUserRegister)
	if !httpx.Bind(c, input, output) {
		return
	}
	output, err := serv.HandleCGUserRegister(c.Request.Context(), input)
	httpx.Response(c, output, err)
}

func HandleCGUserLogin(c *gin.Context) {
	input := new(proto_user.CGUserLogin)
	output := new(proto_user.GCUserLogin)
	if !httpx.Bind(c, input, output) {
		return
	}
	output, err := serv.HandleCGUserLogin(c.Request.Context(), input)
	httpx.Response(c, output, err)
}

func HandleCGMyProfile(c *gin.Context) {
	accessToken := ""
	accessToken = auth.BearerToken(c.GetHeader("Authorization"))
	output, err := serv.HandleCGMyProfile(c.Request.Context(), accessToken)
	httpx.Response(c, output, err)
}

func HandleCGTargetProfile(c *gin.Context) {
	input := new(proto_user.CGTargetProfile)
	output := new(proto_user.GCTargetProfile)
	if !httpx.Bind(c, input, output) {
		return
	}
	accessToken := ""
	accessToken = auth.BearerToken(c.GetHeader("Authorization"))
	output, err := serv.HandleCGTargetProfile(c.Request.Context(), input, accessToken)
	httpx.Response(c, output, err)
}
