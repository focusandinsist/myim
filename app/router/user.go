package router

import (
	"net/http"

	proto_user "myim/api/protobuf/user"
	"myim/internal/httpx"

	"github.com/gin-gonic/gin"
)

func HandleCGUserRegister(c *gin.Context) {
	input := new(proto_user.ReqUserRegister)
	output := new(proto_user.ResUserRegister)
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
	input := new(proto_user.ReqUserLogin)
	output := new(proto_user.ResUserLogin)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = httpx.ErrInvalidRequest.Error()
		httpx.Response(c, output, httpx.ErrInvalidRequest)
		return
	}
	output, err := serv.HandleCGUserLogin(c.Request.Context(), input)
	httpx.Response(c, output, err)
}
