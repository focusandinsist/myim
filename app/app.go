package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myim/app/config"
	"myim/app/dao"
	"myim/app/router"
	"myim/app/service"
)

type App struct {
	config     *config.Config // 应用配置
	dao        *dao.Dao       // PostgreSQL数据访问对象
	httpServer *http.Server   // HTTP服务
}

func New() (*App, error) {
	conf := config.New()
	dao, err := dao.New(conf)
	if err != nil {
		return nil, err
	}
	serv := service.New(conf, dao)
	handler := router.New(serv)
	return &App{
		config: conf,
		dao:    dao,
		httpServer: &http.Server{
			Addr:              conf.Addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
	}, nil
}

func (a *App) Run() error {
	defer a.dao.Close()
	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() {
		serverError <- a.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("run http server: %w", err)
	case <-stopContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.httpServer.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	}
}
