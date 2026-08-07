package service

import (
	"myim/app/config"
	"myim/app/dao"
)

type Service struct {
	config *config.Config     // 服务配置
	dao    dao.UserRepository // 用户数据访问接口
}

func New(c *config.Config, userDAO dao.UserRepository) *Service {
	return &Service{
		config: c,
		dao:    userDAO,
	}
}
