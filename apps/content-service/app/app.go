package app

import (
	"context"
	"errors"
	"myim/apps/content-service/config"
	"myim/apps/content-service/dao"
	"myim/apps/content-service/router"
	"myim/apps/content-service/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	dao    *dao.Dao
	server *http.Server
}

func New() (*App, error) {
	c := config.New()
	d, e := dao.New(c)
	if e != nil {
		return nil, e
	}
	return &App{dao: d, server: &http.Server{Addr: c.Addr, Handler: router.New(service.New(c, d)), ReadHeaderTimeout: 5 * time.Second}}, nil
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
		c, x := context.WithTimeout(context.Background(), 5*time.Second)
		defer x()
		_ = a.server.Shutdown(c)
		return a.dao.Close()
	}
}
