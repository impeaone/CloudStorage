package metrics

import (
	logger2 "CloudStorageProject-FileServer/pkg/logger/logger"
	"CloudStorageProject-FileServer/pkg/tools"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metricsServer struct {
	collector *Collector
	logs      *logger2.Log
	router    http.Handler
	addr      string
}

func newMetricsServer(logger *logger2.Log, collector *Collector) *metricsServer {
	port := tools.GetEnvAsInt("MERICS_SERVER_PORT", 11680)
	ip := tools.GetEnv("MERICS_SERVER_IP", "")

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &metricsServer{
		logs:      logger,
		collector: collector,
		router:    router,
		addr:      fmt.Sprintf("%s:%d", ip, port),
	}
}

func StartMetricsServer(logger *logger2.Log, collector *Collector) {
	server := newMetricsServer(logger, collector)
	go func(logger *logger2.Log, collector *Collector) {
		if err := http.ListenAndServe(server.addr, server.router); err != nil {
			logger.Error("Metrics server start error: "+err.Error(), logger2.GetPlace())
			return

		}
		logger.Info("Metrics server start success on endpoint: "+server.addr, logger2.GetPlace())
	}(logger, collector)
}
