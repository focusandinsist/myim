package router

import (
	proto_social "myim/api/protobuf/social"
	"myim/internal/auth"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

func HandleCGSocialFollow(c *gin.Context) {
	input := new(proto_social.CGSocialFollow)
	output := new(proto_social.GCSocialFollow)
	if !httpx.Bind(c, input, output) {
		return
	}
	output, err := serv.HandleCGSocialFollow(c.Request.Context(), input, auth.BearerToken(c.GetHeader("Authorization")))
	httpx.Response(c, output, err)
}

func HandleCGSocialUnfollow(c *gin.Context) {
	input := new(proto_social.CGSocialUnfollow)
	output := new(proto_social.GCSocialUnfollow)
	if !httpx.Bind(c, input, output) {
		return
	}
	output, err := serv.HandleCGSocialUnfollow(c.Request.Context(), input, auth.BearerToken(c.GetHeader("Authorization")))
	httpx.Response(c, output, err)
}

func HandleCGSocialFollowingList(c *gin.Context) {
	input := new(proto_social.CGSocialFollowingList)
	output := new(proto_social.GCSocialFollowingList)
	if !httpx.Bind(c, input, output) {
		return
	}
	output, err := serv.HandleCGSocialFollowingList(c.Request.Context(), input, auth.BearerToken(c.GetHeader("Authorization")))
	httpx.Response(c, output, err)
}

func HandleCGSocialFollowerList(c *gin.Context) {
	input := new(proto_social.CGSocialFollowerList)
	output := new(proto_social.GCSocialFollowerList)
	if !httpx.Bind(c, input, output) {
		return
	}
	output, err := serv.HandleCGSocialFollowerList(c.Request.Context(), input, auth.BearerToken(c.GetHeader("Authorization")))
	httpx.Response(c, output, err)
}

func HandleCGSocialFollowCheck(c *gin.Context) {
	input := new(proto_social.CGSocialFollowCheck)
	output := new(proto_social.GCSocialFollowCheck)
	if !httpx.Bind(c, input, output) {
		return
	}
	output, err := serv.HandleCGSocialFollowCheck(c.Request.Context(), input, auth.BearerToken(c.GetHeader("Authorization")))
	httpx.Response(c, output, err)
}
