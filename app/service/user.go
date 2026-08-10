package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	proto_user "myim/api/protobuf/user"
	"myim/app/dao"
	"myim/app/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput       = errors.New("invalid input")                   // 请求字段缺失或格式不符合业务要求
	ErrUserNameExists     = errors.New("user name already exists")        // 注册用户名已被占用
	ErrInvalidCredentials = errors.New("invalid user name or password")   // 登录用户名不存在或密码错误
	ErrInvalidAccessToken = errors.New("invalid or expired access token") // 访问令牌无效、过期或所属用户不存在
	ErrUserNotFound       = errors.New("user not found")                  // 目标用户不存在
)

const defaultTokenExpiry = 24 * time.Hour

// AccessTokenClaims 表示 myim 访问令牌中保存的身份和有效期信息。
type AccessTokenClaims struct {
	UserID    string `json:"user_id"`   // 用户ID
	UserName  string `json:"user_name"` // 登录用户名
	IssuedAt  int64  `json:"iat"`       // 签发时间，Unix秒
	ExpiresAt int64  `json:"exp"`       // 过期时间，Unix秒
}

func (s *Service) HandleCGUserRegister(ctx context.Context, input *proto_user.ReqUserRegister) (output *proto_user.ResUserRegister, err error) {
	output = new(proto_user.ResUserRegister)
	if input == nil {
		err = fmt.Errorf("%w: request is required", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}

	userName := strings.TrimSpace(input.GetUserName())
	password := input.GetPassword()
	if utf8.RuneCountInString(userName) < 3 || utf8.RuneCountInString(userName) > 32 {
		err = fmt.Errorf("%w: user_name must contain 3 to 32 characters", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}
	for _, char := range userName {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			err = fmt.Errorf("%w: user_name may only contain letters, numbers, and underscores", ErrInvalidInput)
			output.ErrorCode = http.StatusBadRequest
			output.ErrorMsg = err.Error()
			return output, err
		}
	}

	if utf8.RuneCountInString(password) < 8 || len(password) > 72 {
		err = fmt.Errorf("%w: password must contain 8 to 72 characters", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}
	var hasLetter, hasNumber bool
	for _, char := range password {
		if unicode.IsSpace(char) {
			err = fmt.Errorf("%w: password must not contain whitespace", ErrInvalidInput)
			output.ErrorCode = http.StatusBadRequest
			output.ErrorMsg = err.Error()
			return output, err
		}
		hasLetter = hasLetter || unicode.IsLetter(char)
		hasNumber = hasNumber || unicode.IsDigit(char)
	}
	if !hasLetter || !hasNumber {
		err = fmt.Errorf("%w: password must contain at least one letter and one number", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}

	exists, err := s.dao.UserNameExists(ctx, userName)
	if err != nil {
		err = fmt.Errorf("check user name: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	if exists {
		output.ErrorCode = http.StatusConflict
		output.ErrorMsg = ErrUserNameExists.Error()
		return output, ErrUserNameExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		err = fmt.Errorf("hash password: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}

	now := time.Now().Unix()
	user := &model.User{
		UserID:       uuid.NewString(),
		UserName:     userName,
		Password:     string(passwordHash),
		Nickname:     userName,
		Status:       0,
		RegisterTime: now,
		UpdatedTime:  now,
	}
	if err = s.dao.CreateUser(ctx, user); err != nil {
		if errors.Is(err, dao.ErrUserNameExists) {
			output.ErrorCode = http.StatusConflict
			output.ErrorMsg = ErrUserNameExists.Error()
			return output, ErrUserNameExists
		}
		err = fmt.Errorf("create user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}

	output.ErrorMsg = "ok"
	output.UserId = user.UserID
	return output, nil
}

func (s *Service) HandleCGUserLogin(ctx context.Context, input *proto_user.ReqUserLogin) (output *proto_user.ResUserLogin, err error) {
	output = new(proto_user.ResUserLogin)
	if input == nil {
		err = fmt.Errorf("%w: request is required", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}

	userName := strings.TrimSpace(input.GetUserName())
	password := input.GetPassword()
	if userName == "" || password == "" {
		err = fmt.Errorf("%w: user_name and password are required", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}

	user, err := s.dao.GetUserByUserName(ctx, userName)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			output.ErrorCode = http.StatusUnauthorized
			output.ErrorMsg = ErrInvalidCredentials.Error()
			return output, ErrInvalidCredentials
		}
		err = fmt.Errorf("get user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = ErrInvalidCredentials.Error()
		return output, ErrInvalidCredentials
	}

	tokenExpiry := s.config.TokenExpires
	if tokenExpiry <= 0 {
		tokenExpiry = defaultTokenExpiry
	}
	now := time.Now()
	expiresAt := now.Add(tokenExpiry).Unix()
	token, err := generateAccessToken(AccessTokenClaims{
		UserID: user.UserID, UserName: user.UserName, IssuedAt: now.Unix(), ExpiresAt: expiresAt,
	}, s.config.JWTSecret)
	if err != nil {
		err = fmt.Errorf("generate access token: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}
	if err = s.dao.UpdateLastLoginTime(ctx, user.UserID, now.Unix()); err != nil {
		err = fmt.Errorf("update last login time: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}

	output.ErrorMsg = "ok"
	output.UserId = user.UserID
	output.AccessToken = token
	output.ExpiresAt = expiresAt
	return output, nil
}

func (s *Service) HandleCGMyProfile(ctx context.Context, accessToken string) (output *proto_user.ResMyProfile, err error) {
	output = new(proto_user.ResMyProfile)
	claims, err := ValidateAccessToken(accessToken, s.config.JWTSecret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = ErrInvalidAccessToken.Error()
		return output, ErrInvalidAccessToken
	}

	user, err := s.dao.GetUserByUserID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			output.ErrorCode = http.StatusUnauthorized
			output.ErrorMsg = ErrInvalidAccessToken.Error()
			return output, ErrInvalidAccessToken
		}
		err = fmt.Errorf("get current user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}

	output.ErrorMsg = "ok"
	output.UserId = user.UserID
	output.UserName = user.UserName
	output.Phone = user.Phone
	output.Email = user.Email
	output.Nickname = user.Nickname
	output.Avatar = user.Avatar
	output.Bio = user.Bio
	output.Gender = user.Gender
	output.Birthday = user.Birthday
	output.Region = user.Region
	output.Status = user.Status
	output.RegisterTime = user.RegisterTime
	output.LastLoginTime = user.LastLoginTime
	output.UpdatedTime = user.UpdatedTime
	return output, nil
}

func (s *Service) HandleCGTargetProfile(ctx context.Context, input *proto_user.ReqTargetProfile, accessToken string) (output *proto_user.ResTargetProfile, err error) {
	output = new(proto_user.ResTargetProfile)
	if _, err = ValidateAccessToken(accessToken, s.config.JWTSecret); err != nil {
		output.ErrorCode = http.StatusUnauthorized
		output.ErrorMsg = ErrInvalidAccessToken.Error()
		return output, ErrInvalidAccessToken
	}
	if input == nil || uuid.Validate(input.GetUserId()) != nil {
		err = fmt.Errorf("%w: valid user_id is required", ErrInvalidInput)
		output.ErrorCode = http.StatusBadRequest
		output.ErrorMsg = err.Error()
		return output, err
	}

	user, err := s.dao.GetUserByUserID(ctx, input.GetUserId())
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			output.ErrorCode = http.StatusNotFound
			output.ErrorMsg = ErrUserNotFound.Error()
			return output, ErrUserNotFound
		}
		err = fmt.Errorf("get other user: %w", err)
		output.ErrorCode = http.StatusInternalServerError
		output.ErrorMsg = err.Error()
		return output, err
	}

	output.ErrorMsg = "ok"
	output.UserId = user.UserID
	output.Nickname = user.Nickname
	output.Avatar = user.Avatar
	output.Bio = user.Bio
	output.Gender = user.Gender
	output.Birthday = user.Birthday
	output.Region = user.Region
	return output, nil
}

func generateAccessToken(claims AccessTokenClaims, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("empty JWT secret")
	}
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	signature := signToken(unsigned, secret)
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func ValidateAccessToken(token, secret string) (*AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || secret == "" {
		return nil, errors.New("invalid access token")
	}
	expected := signToken(parts[0]+"."+parts[1], secret)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actual, expected) {
		return nil, errors.New("invalid access token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid access token")
	}
	claims := new(AccessTokenClaims)
	if err := json.Unmarshal(payload, claims); err != nil || claims.UserID == "" || claims.ExpiresAt <= time.Now().Unix() {
		return nil, errors.New("invalid or expired access token")
	}
	return claims, nil
}

func signToken(unsigned, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return mac.Sum(nil)
}
