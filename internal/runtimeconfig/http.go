package runtimeconfig

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
}

func DefaultHTTP() HTTPConfig {
	return HTTPConfig{
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   5 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

func (c HTTPConfig) Validate(addr string, websocket bool) error {
	if strings.TrimSpace(addr) == "" {
		return errors.New("HTTP listen address is required")
	}
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid HTTP listen address: %w", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return errors.New("HTTP listen port must be between 1 and 65535")
	}
	if c.ReadHeaderTimeout <= 0 || c.IdleTimeout <= 0 || c.ShutdownTimeout <= 0 || c.MaxHeaderBytes <= 0 {
		return errors.New("HTTP header, idle, shutdown and header-size limits must be positive")
	}
	if !websocket && (c.ReadTimeout <= 0 || c.WriteTimeout <= 0) {
		return errors.New("HTTP read and write timeouts must be positive")
	}
	return nil
}

func (c HTTPConfig) Server(addr string, handler http.Handler, websocket bool) *http.Server {
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: c.ReadHeaderTimeout,
		IdleTimeout:       c.IdleTimeout,
		MaxHeaderBytes:    c.MaxHeaderBytes,
	}
	if !websocket {
		server.ReadTimeout = c.ReadTimeout
		server.WriteTimeout = c.WriteTimeout
	}
	return server
}

func ShutdownHTTP(ctx context.Context, server *http.Server) error {
	if err := server.Shutdown(ctx); err != nil {
		closeErr := server.Close()
		return errors.Join(fmt.Errorf("graceful HTTP shutdown: %w", err), closeErr)
	}
	return nil
}
