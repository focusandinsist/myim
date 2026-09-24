package model

type User struct {
	UserID        string `json:"user_id"`         // 用户ID，主键
	UserName      string `json:"user_name"`       // 登录用户名
	Password      string `json:"-"`               // bcrypt密码哈希
	Phone         string `json:"phone"`           // 手机号
	Email         string `json:"email"`           // 邮箱
	Nickname      string `json:"nickname"`        // 社交昵称
	Avatar        string `json:"avatar"`          // 头像地址
	Bio           string `json:"bio"`             // 个人简介
	Gender        int32  `json:"gender"`          // 性别：0未知，1男，2女，3其他
	Birthday      int64  `json:"birthday"`        // 生日，Unix秒
	Region        string `json:"region"`          // 地区
	Status        int32  `json:"status"`          // 状态：0正常，1禁用，2注销
	RegisterTime  int64  `json:"register_time"`   // 注册时间
	LastLoginTime int64  `json:"last_login_time"` // 最后登录时间
	UpdatedTime   int64  `json:"updated_time"`    // 最后更新时间
}
