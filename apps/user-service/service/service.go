package service

import (
	"context"
	"errors"

	"myim/apps/user-service/config"
	"myim/apps/user-service/dao"
)

func (s *Service) Health(ctx context.Context) error {
	if s == nil || s.dao == nil {
		return errors.New("user database is unavailable")
	}
	return s.dao.Ping(ctx)
}

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
