package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"myim/apps/message-service/config"
	"myim/apps/message-service/dao"
	"myim/apps/message-service/router"
	"myim/apps/message-service/service"
	"myim/internal/runtimeconfig"
)

type App struct {
	config     *config.Config   // message service配置
	dao        *dao.Dao         // message PostgreSQL数据访问对象
	service    *service.Service // message业务服务
	httpServer *http.Server     // HTTP和WebSocket服务
	stopOnce   sync.Once        // 保证应用资源只关闭一次
	stopErr    error            // 保存资源关闭结果
}

func New() (*App, error) {
	conf := config.New()
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("validate message service config: %w", err)
	}
	dao, err := dao.New(conf)
	if err != nil {
		return nil, err
	}
	serv := service.New(conf, dao)
	handler := router.New(conf, serv)
	return &App{
		config:     conf,
		dao:        dao,
		service:    serv,
		httpServer: conf.HTTP.Server(conf.Addr, handler, true),
	}, nil
}

func (a *App) Run() error {
	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() {
		serverError <- a.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		shutdownContext, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
		defer cancel()
		stopErr := a.Stop(shutdownContext)
		if !errors.Is(err, http.ErrServerClosed) {
			err = fmt.Errorf("run message server: %w", err)
			if stopErr != nil {
				return errors.Join(err, stopErr)
			}
			return err
		}
		return stopErr
	case <-stopContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
		defer cancel()
		return a.Stop(shutdownContext)
	}
}

func (a *App) Stop(ctx context.Context) error {
	a.stopOnce.Do(func() {
		var stopErrors []error
		a.service.Stop()
		if err := runtimeconfig.ShutdownHTTP(ctx, a.httpServer); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("shutdown message server: %w", err))
		}
		if err := router.Wait(ctx); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("wait for message HTTP handlers: %w", err))
		} else if err := a.dao.Close(); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("close message dao: %w", err))
		}
		a.stopErr = errors.Join(stopErrors...)
	})
	return a.stopErr
}
