package service

import (
	"myim/apps/user-service/config"
	"myim/apps/user-service/dao"
)

type Service struct {
	config *config.Config // 服务配置
	dao    *dao.Dao       // PostgreSQL数据访问对象
}

func New(c *config.Config, dao *dao.Dao) *Service {
	return &Service{
		config: c,
		dao:    dao,
	}
}
