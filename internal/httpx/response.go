package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Response(c *gin.Context, obj any, err error) {
	statusCode := http.StatusOK
	if err != nil {
		statusCode = http.StatusInternalServerError
		protoMessage, ok := obj.(proto.Message)
		if ok && protoMessage != nil {
			reflection := protoMessage.ProtoReflect()
			fields := reflection.Descriptor().Fields()
			if field := fields.ByName("error_code"); field != nil && field.Kind() == protoreflect.Int32Kind {
				code := int(reflection.Get(field).Int())
				if code >= http.StatusBadRequest && code <= 599 {
					statusCode = code
				}
				reflection.Set(field, protoreflect.ValueOfInt32(int32(statusCode)))
			}
			if field := fields.ByName("error_msg"); field != nil && field.Kind() == protoreflect.StringKind {
				message := err.Error()
				if statusCode >= http.StatusInternalServerError {
					message = ErrInternalServer.Error()
				}
				reflection.Set(field, protoreflect.ValueOfString(message))
			}
		}
	}

	protoMessage, isProto := obj.(proto.Message)
	accept := strings.ToLower(c.GetHeader("Accept"))
	contentType := strings.ToLower(c.ContentType())
	useProtobuf := strings.Contains(accept, "application/x-protobuf") ||
		strings.Contains(accept, "application/protobuf") ||
		strings.Contains(contentType, "application/x-protobuf") ||
		strings.Contains(contentType, "application/protobuf")
	if isProto && useProtobuf {
		c.ProtoBuf(statusCode, protoMessage)
		return
	}
	if isProto {
		data, marshalErr := (protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}).Marshal(protoMessage)
		if marshalErr == nil {
			c.Data(statusCode, "application/json; charset=utf-8", data)
			return
		}
	}
	c.JSON(statusCode, obj)
}
