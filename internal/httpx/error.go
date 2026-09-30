package httpx

import "errors"

var (
	ErrInvalidRequest      = errors.New("invalid input: invalid request") // 请求体格式错误或无法按声明协议解析
	ErrRequestBodyTooLarge = errors.New("request body too large")         // 请求体超过统一大小限制
	ErrInternalServer      = errors.New("internal server error")          // 未向客户端暴露细节的服务端错误
)
