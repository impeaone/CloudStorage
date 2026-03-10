package metrics

import (
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	collector *Collector
	logger    *slog.Logger
	router    http.Handler
	addr      string
	exitChan  chan struct{}
	cons      *sync.WaitGroup
}

func NewMetricsServer(ctx context.Context) *MetricsServer {
	conf := ctx.Value("config").(*config.Config)
	logger := ctx.Value("logger").(*slog.Logger)

	router := http.NewServeMux()
	exitChan := make(chan struct{})
	cons := new(sync.WaitGroup)

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	routerWithShutdown := ShutdownMiddleware(exitChan, cons, router)
	routerWithLogger := Logger(logger, routerWithShutdown)
	return &MetricsServer{
		logger:   logger,
		router:   routerWithLogger,
		addr:     conf.MetricsAddr(),
		exitChan: exitChan,
		cons:     cons,
	}
}

func (m *MetricsServer) StartMetricsServer() error {
	if err := http.ListenAndServe(m.addr, m.router); err != nil {
		return err
	}
	return nil
}

func (m *MetricsServer) Close(ctx context.Context) error {
	close(m.exitChan)

	finished := make(chan struct{})
	go func() {
		m.cons.Wait()
		close(finished)
	}()

	select {
	case <-finished:
		// Все операции завершилсь
		return nil
	case <-ctx.Done():
		// Время истекло
		return ctx.Err()
	}
}

// ShutdownMiddleware - middleware проверяющий закрыт ли канал, для graceful shutdown
func ShutdownMiddleware(exitChan chan struct{}, conns *sync.WaitGroup, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-exitChan:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "service_unavailable",
				"message": "Service is shutting down",
			})
			return
		default:
			conns.Add(1)
			defer conns.Done()
			next.ServeHTTP(w, r)
		}
	})
}

// Logger - middleware, логги
func Logger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "static") {
			logger.Info("metrics logger", "time", time.Now().String(), "url", r.URL, "client",
				r.RemoteAddr, "method", r.Method, "place", tools.GetPlace())
		}
		next.ServeHTTP(w, r)
	})
}
