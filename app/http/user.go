package http

import (
	"fmt"
	nethttp "net/http"

	proto_user "myim/api/protobuf/user"
	"myim/app/service"
	"myim/internal/http"

	"github.com/gin-gonic/gin"
)

func HandleCGUserRegister(c *gin.Context) {
	input := new(proto_user.ReqUserRegister)
	output := new(proto_user.ResUserRegister)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = nethttp.StatusBadRequest
		output.ErrorMsg = "invalid input: invalid request"
		http.Response(c, output, fmt.Errorf("%w: invalid request", service.ErrInvalidInput))
		return
	}
	output, err := serv.HandleCGUserRegister(c.Request.Context(), input)
	http.Response(c, output, err)
}

func HandleCGUserLogin(c *gin.Context) {
	input := new(proto_user.ReqUserLogin)
	output := new(proto_user.ResUserLogin)
	if err := c.ShouldBind(input); err != nil {
		output.ErrorCode = nethttp.StatusBadRequest
		output.ErrorMsg = "invalid input: invalid request"
		http.Response(c, output, fmt.Errorf("%w: invalid request", service.ErrInvalidInput))
		return
	}
	output, err := serv.HandleCGUserLogin(c.Request.Context(), input)
	http.Response(c, output, err)
}
