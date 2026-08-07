package http

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myim/app/service"

	"github.com/gin-gonic/gin"
)

var serv *service.Service

func Init(s *service.Service) *gin.Engine {
	serv = s
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) { c.Status(nethttp.StatusNoContent) })
	router.POST("/user-register", HandleCGUserRegister)
	router.POST("/user-login", HandleCGUserLogin)
	return router
}

func Run(addr string, handler nethttp.Handler) error {
	server := &nethttp.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() {
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if errors.Is(err, nethttp.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("run http server: %w", err)
	case <-stopContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	}
}
