package service

import (
	"myim/app/config"
	"myim/app/dao"
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
