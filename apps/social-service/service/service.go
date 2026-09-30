package service

import (
	"context"
	"errors"

	"myim/apps/social-service/config"
	"myim/apps/social-service/dao"
)

func (s *Service) Health(ctx context.Context) error {
	if s == nil || s.dao == nil {
		return errors.New("social database is unavailable")
	}
	return s.dao.Ping(ctx)
}

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
