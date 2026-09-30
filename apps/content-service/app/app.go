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

	"myim/apps/content-service/config"
	"myim/apps/content-service/dao"
	"myim/apps/content-service/router"
	"myim/apps/content-service/service"
	"myim/internal/runtimeconfig"
)

type App struct {
	config       *config.Config     // Content服务配置
	dao          *dao.Dao           // Content PostgreSQL数据访问对象
	service      *service.Service   // Content业务服务
	server       *http.Server       // Content HTTP服务
	stopOnce     sync.Once          // 保证应用资源只关闭一次
	stopErr      error              // 保存资源关闭结果
	workerCancel context.CancelFunc // 停止Outbox worker
	workerDone   chan struct{}      // 等待Outbox worker退出
}

func New() (*App, error) {
	c := config.New()
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("validate content service config: %w", err)
	}
	d, e := dao.New(c)
	if e != nil {
		return nil, e
	}
	serv := service.New(c, d)
	return &App{
		config:  c,
		dao:     d,
		service: serv,
		server:  c.HTTP.Server(c.Addr, router.New(serv), false),
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	workerCtx, cancelWorker := context.WithCancel(ctx)
	a.workerCancel = cancelWorker
	a.workerDone = make(chan struct{})
	go func() {
		defer close(a.workerDone)
		a.service.RunOutbox(workerCtx)
	}()
	ch := make(chan error, 1)
	go func() { ch <- a.server.ListenAndServe() }()
	select {
	case e := <-ch:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
		defer cancel()
		stopErr := a.Stop(shutdownCtx)
		if errors.Is(e, http.ErrServerClosed) {
			return stopErr
		}
		if stopErr != nil {
			return errors.Join(fmt.Errorf("run content http server: %w", e), stopErr)
		}
		return fmt.Errorf("run content http server: %w", e)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
		defer cancel()
		return a.Stop(shutdownCtx)
	}
}

func (a *App) Stop(ctx context.Context) error {
	a.stopOnce.Do(func() {
		var stopErrors []error
		handlersStopped := true
		if err := runtimeconfig.ShutdownHTTP(ctx, a.server); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("shutdown content http server: %w", err))
		}
		if err := router.Wait(ctx); err != nil {
			handlersStopped = false
			stopErrors = append(stopErrors, fmt.Errorf("wait for content HTTP handlers: %w", err))
		}
		if a.workerCancel != nil {
			a.workerCancel()
		}
		workerStopped := a.workerDone == nil
		if a.workerDone != nil {
			select {
			case <-a.workerDone:
				workerStopped = true
			case <-ctx.Done():
				stopErrors = append(stopErrors, fmt.Errorf("stop content outbox worker: %w", ctx.Err()))
			}
		}
		a.service.Stop()
		if !workerStopped && a.workerDone != nil {
			select {
			case <-a.workerDone:
				workerStopped = true
			case <-time.After(time.Second):
				stopErrors = append(stopErrors, errors.New("content outbox worker did not stop after publisher close"))
			}
		}
		if !workerStopped {
			a.stopErr = errors.Join(stopErrors...)
			return
		}
		if handlersStopped {
			if err := a.dao.Close(); err != nil {
				stopErrors = append(stopErrors, fmt.Errorf("close content dao: %w", err))
			}
		} else {
			stopErrors = append(stopErrors, errors.New("content dao left open while HTTP handlers are active"))
		}
		a.stopErr = errors.Join(stopErrors...)
	})
	return a.stopErr
}
