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
	"time"

	"myim/apps/social-service/config"
	"myim/apps/social-service/dao"
	"myim/apps/social-service/router"
	"myim/apps/social-service/service"
)

type App struct {
	config     *config.Config // Social应用配置
	dao        *dao.Dao       // Social PostgreSQL数据访问对象
	httpServer *http.Server   // Social HTTP服务
	stopOnce   sync.Once      // 保证应用资源只关闭一次
	stopErr    error          // 保存资源关闭结果
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
	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() {
		serverError <- a.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stopErr := a.Stop(shutdownContext)
		if !errors.Is(err, http.ErrServerClosed) {
			err = fmt.Errorf("run social http server: %w", err)
			if stopErr != nil {
				return errors.Join(err, stopErr)
			}
			return err
		}
		return stopErr
	case <-stopContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return a.Stop(shutdownContext)
	}
}

func (a *App) Stop(ctx context.Context) error {
	a.stopOnce.Do(func() {
		var stopErrors []error
		if err := a.httpServer.Shutdown(ctx); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("shutdown social http server: %w", err))
		}
		if err := a.dao.Close(); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("close social dao: %w", err))
		}
		a.stopErr = errors.Join(stopErrors...)
	})
	return a.stopErr
}
