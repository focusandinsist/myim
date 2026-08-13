package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAccessToken = errors.New("invalid access token") // 访问令牌缺失、格式错误或签名无效
	ErrExpiredAccessToken = errors.New("expired access token") // 访问令牌已超过有效期
)

// AccessTokenClaims 表示 message service 从访问令牌读取的用户身份和有效期。
type AccessTokenClaims struct {
	UserID    string `json:"user_id"`   // 用户ID
	UserName  string `json:"user_name"` // 登录用户名
	IssuedAt  int64  `json:"iat"`       // 签发时间，Unix秒
	ExpiresAt int64  `json:"exp"`       // 过期时间，Unix秒
}

func ValidateAccessToken(token, secret string) (*AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || secret == "" {
		return nil, ErrInvalidAccessToken
	}
	expectedMAC := hmac.New(sha256.New, []byte(secret))
	expectedMAC.Write([]byte(parts[0] + "." + parts[1]))
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actual, expectedMAC.Sum(nil)) {
		return nil, ErrInvalidAccessToken
	}
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidAccessToken
	}
	header := make(map[string]string)
	if err := json.Unmarshal(headerData, &header); err != nil || header["alg"] != "HS256" {
		return nil, ErrInvalidAccessToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidAccessToken
	}
	claims := new(AccessTokenClaims)
	if err := json.Unmarshal(payload, claims); err != nil || uuid.Validate(claims.UserID) != nil || claims.ExpiresAt <= 0 {
		return nil, ErrInvalidAccessToken
	}
	if claims.ExpiresAt <= time.Now().Unix() {
		return nil, ErrExpiredAccessToken
	}
	return claims, nil
}
