package metrics

import (
	logger2 "CloudStorageProject-FileServer/pkg/logger/logger"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	collector *Collector
	logs      *logger2.Log
	router    http.Handler
	addr      string
	exitChan  chan struct{}
	cons      *sync.WaitGroup
}

func newMetricsServer(logger *logger2.Log, collector *Collector) *MetricsServer {
	port := tools.GetEnvAsInt("METRICS_SERVER_PORT", 11680)
	ip := tools.GetEnv("METRICS_SERVER_IP", "")

	router := http.NewServeMux()
	exitChan := make(chan struct{})
	cons := new(sync.WaitGroup)

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	routerWithShutdown := ShutdownMiddleware(exitChan, cons, router)
	routerWithLogger := Logger(logger, routerWithShutdown)
	return &MetricsServer{
		logs:      logger,
		collector: collector,
		router:    routerWithLogger,
		addr:      fmt.Sprintf("%s:%d", ip, port),
		exitChan:  exitChan,
		cons:      cons,
	}
}

func StartMetricsServer(logger *logger2.Log, collector *Collector) *MetricsServer {
	server := newMetricsServer(logger, collector)
	go func(logger *logger2.Log, collector *Collector) {
		if err := http.ListenAndServe(server.addr, server.router); err != nil {
			logger.Error("Metrics server start error: "+err.Error(), logger2.GetPlace())
			return
		}
		logger.Info("Metrics server start success on endpoint: "+server.addr, logger2.GetPlace())
	}(logger, collector)
	return server
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
		conns.Add(1)
		defer conns.Done()
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
			next.ServeHTTP(w, r)
		}
	})
}

// Logger - middleware, логги
func Logger(logs *logger2.Log, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "static") {
			logs.Info(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v",
				r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		}
		next.ServeHTTP(w, r)
	})
}
