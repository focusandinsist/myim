package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const DevelopmentSecret = "myim-development-secret"

var (
	ErrInvalidAccessToken = errors.New("invalid access token") // 访问令牌缺失、格式错误或签名无效
	ErrExpiredAccessToken = errors.New("expired access token") // 访问令牌已超过有效期
)

type Config struct {
	Secret       string        // JWT 签名密钥
	TokenExpires time.Duration // 访问令牌有效期
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Secret) == "" {
		return errors.New("JWT secret is required")
	}
	if c.Secret == DevelopmentSecret {
		return errors.New("development JWT secret must not be used by a service")
	}
	if c.TokenExpires <= 0 {
		return errors.New("token expiry must be positive")
	}
	return nil
}

type Claims struct {
	UserID    string `json:"user_id"`   // 用户 ID
	UserName  string `json:"user_name"` // 登录用户名
	IssuedAt  int64  `json:"iat"`       // 签发时间，Unix 秒
	ExpiresAt int64  `json:"exp"`       // 过期时间，Unix 秒
}

func Issue(claims Claims, secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("JWT secret is required")
	}
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal token header: %w", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal token claims: %w", err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(sign(unsigned, secret)), nil
}

func Validate(token, secret string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || strings.TrimSpace(secret) == "" {
		return nil, ErrInvalidAccessToken
	}
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidAccessToken
	}
	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	if err := json.Unmarshal(headerData, &header); err != nil || header.Algorithm != "HS256" || header.Type != "JWT" {
		return nil, ErrInvalidAccessToken
	}
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actual, sign(parts[0]+"."+parts[1], secret)) {
		return nil, ErrInvalidAccessToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidAccessToken
	}
	claims := new(Claims)
	if err := json.Unmarshal(payload, claims); err != nil || claims.UserID == "" || claims.ExpiresAt <= 0 {
		return nil, ErrInvalidAccessToken
	}
	if claims.ExpiresAt <= time.Now().Unix() {
		return nil, ErrExpiredAccessToken
	}
	return claims, nil
}

func BearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func RequireBearerToken(header string) (string, error) {
	token := BearerToken(header)
	if token == "" {
		return "", ErrInvalidAccessToken
	}
	return token, nil
}

func sign(unsigned, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(unsigned))
	return mac.Sum(nil)
}
