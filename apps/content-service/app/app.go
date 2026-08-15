package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myim/apps/content-service/config"
	"myim/apps/content-service/dao"
	"myim/apps/content-service/router"
	"myim/apps/content-service/service"
)

type App struct {
	dao     *dao.Dao         // Content PostgreSQL数据访问对象
	service *service.Service // Content业务服务
	server  *http.Server     // Content HTTP服务
}

func New() (*App, error) {
	c := config.New()
	d, e := dao.New(c)
	if e != nil {
		return nil, e
	}
	serv := service.New(c, d)
	return &App{
		dao:     d,
		service: serv,
		server: &http.Server{
			Addr:              c.Addr,
			Handler:           router.New(serv),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ch := make(chan error, 1)
	go func() { ch <- a.server.ListenAndServe() }()
	select {
	case e := <-ch:
		if errors.Is(e, http.ErrServerClosed) {
			return nil
		}
		return e
	case <-ctx.Done():
		a.service.Stop()
		c, x := context.WithTimeout(context.Background(), 5*time.Second)
		defer x()
		_ = a.server.Shutdown(c)
		return a.dao.Close()
	}
}
