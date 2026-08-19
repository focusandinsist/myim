package service

import (
	"myim/apps/social-service/config"
	"myim/apps/social-service/dao"
)

type Service struct {
	config *config.Config // Social服务配置
	dao    *dao.Dao       // PostgreSQL数据访问对象
}

func New(c *config.Config, dao *dao.Dao) *Service {
	return &Service{
		config: c,
		dao:    dao,
	}
}
