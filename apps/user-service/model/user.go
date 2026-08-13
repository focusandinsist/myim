package model

type User struct {
	UserID        string `json:"user_id"         gorm:"column:user_id;primaryKey;type:varchar(36)"`  // 用户ID，主键
	UserName      string `json:"user_name"       gorm:"column:user_name;type:varchar(32);index"`     // 登录用户名
	Password      string `json:"-"               gorm:"column:password;type:varchar(60)"`            // bcrypt密码哈希
	Phone         string `json:"phone"           gorm:"column:phone;type:varchar(32);index"`         // 手机号
	Email         string `json:"email"           gorm:"column:email;type:varchar(128);index"`        // 邮箱
	Nickname      string `json:"nickname"        gorm:"column:nickname;type:varchar(64)"`            // 社交昵称
	Avatar        string `json:"avatar"          gorm:"column:avatar;type:varchar(512)"`             // 头像地址
	Bio           string `json:"bio"             gorm:"column:bio;type:varchar(256)"`                // 个人简介
	Gender        int32  `json:"gender"          gorm:"column:gender;type:smallint;default:0"`       // 性别：0未知，1男，2女，3其他
	Birthday      int64  `json:"birthday"        gorm:"column:birthday;default:0"`                   // 生日，Unix秒
	Region        string `json:"region"          gorm:"column:region;type:varchar(128)"`             // 地区
	Status        int32  `json:"status"          gorm:"column:status;type:smallint;default:0;index"` // 状态：0正常，1禁用，2注销
	RegisterTime  int64  `json:"register_time"   gorm:"column:register_time"`                        // 注册时间
	LastLoginTime int64  `json:"last_login_time" gorm:"column:last_login_time;default:0"`            // 最后登录时间
	UpdatedTime   int64  `json:"updated_time"    gorm:"column:updated_time"`                         // 最后更新时间
}

func (User) TableName() string {
	return "users"
}
