package observability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
	MaxRequestBytes = 1 << 20
)

type requestIDContextKey struct{}
type bodyLimitContextKey struct{}

type limitedBody struct {
	io.ReadCloser
	remaining int64
	atLimit   bool
	exceeded  bool
}

func (b *limitedBody) Read(p []byte) (int, error) {
	wasAtLimit := b.atLimit
	n, err := b.ReadCloser.Read(p)
	b.remaining -= int64(n)
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) || (wasAtLimit && n > 0) {
		b.exceeded = true
	}
	if b.remaining <= 0 && err == nil {
		b.atLimit = true
	}
	return n, err
}

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type requestMetric struct {
	method      string
	route       string
	status      int
	count       uint64
	durationSum time.Duration
}

type HTTP struct {
	mu       sync.Mutex
	metrics  map[string]requestMetric
	logger   *slog.Logger
	requests sync.WaitGroup
}

func NewHTTP() *HTTP {
	return &HTTP{
		metrics: make(map[string]requestMetric),
		logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func (h *HTTP) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.requests.Add(1)
		defer h.requests.Done()
		requestID := c.GetHeader(RequestIDHeader)
		if !requestIDPattern.MatchString(requestID) {
			requestID = uuid.NewString()
		}
		body := &limitedBody{ReadCloser: http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestBytes), remaining: MaxRequestBytes}
		ctx := context.WithValue(c.Request.Context(), requestIDContextKey{}, requestID)
		ctx = context.WithValue(ctx, bodyLimitContextKey{}, body)
		c.Request = c.Request.WithContext(ctx)
		c.Header(RequestIDHeader, requestID)
		c.Request.Body = body
		startedAt := time.Now()
		if c.Request.ContentLength > MaxRequestBytes {
			c.AbortWithStatus(http.StatusRequestEntityTooLarge)
		}
		c.Next()
		elapsed := time.Since(startedAt)
		status := c.Writer.Status()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		key := metricKey(c.Request.Method, route, status)
		h.mu.Lock()
		metric := h.metrics[key]
		metric.method = c.Request.Method
		metric.route = route
		metric.status = status
		metric.count++
		metric.durationSum += elapsed
		h.metrics[key] = metric
		h.mu.Unlock()
		h.logger.InfoContext(c.Request.Context(), "http request",
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("route", route),
			slog.Int("status", status),
			slog.Duration("duration", elapsed),
		)
	}
}

func BodyLimitExceeded(ctx context.Context) bool {
	body, _ := ctx.Value(bodyLimitContextKey{}).(*limitedBody)
	return body != nil && body.exceeded
}

// Wait blocks until all requests that entered this middleware have returned.
func (h *HTTP) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		h.requests.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Health(check func(context.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if check == nil || check(ctx) != nil {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (h *HTTP) Metrics(c *gin.Context) {
	h.mu.Lock()
	metrics := make([]requestMetric, 0, len(h.metrics))
	for _, metric := range h.metrics {
		metrics = append(metrics, metric)
	}
	h.mu.Unlock()
	sort.Slice(metrics, func(i, j int) bool {
		return metricKey(metrics[i].method, metrics[i].route, metrics[i].status) < metricKey(metrics[j].method, metrics[j].route, metrics[j].status)
	})
	var body strings.Builder
	body.WriteString("# TYPE myim_http_requests_total counter\n")
	body.WriteString("# TYPE myim_http_request_duration_seconds_sum counter\n")
	for _, metric := range metrics {
		labels := fmt.Sprintf(`method="%s",route="%s",status="%d"`, escapeLabel(metric.method), escapeLabel(metric.route), metric.status)
		fmt.Fprintf(&body, "myim_http_requests_total{%s} %d\n", labels, metric.count)
		fmt.Fprintf(&body, "myim_http_request_duration_seconds_sum{%s} %.9f\n", labels, metric.durationSum.Seconds())
	}
	c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(body.String()))
}

func metricKey(method, route string, status int) string {
	return fmt.Sprintf("%s\x00%s\x00%d", method, route, status)
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
