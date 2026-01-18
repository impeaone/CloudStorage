package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector централизованный сборщик всех метрик
type Collector struct {
	HTTP     *HTTPMetrics
	Postgres *PostgresMetrics
	Minio    *MinIOMetrics
	Redis    *RedisMetrics
	Custom   *CustomMetrics
}

// HTTPMetrics метрики HTTP запросов
type HTTPMetrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestSize      *prometheus.HistogramVec
	ResponseSize     *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge
}

// PostgresMetrics метрики базы данных PostgreSQL
type PostgresMetrics struct {
	QueryDuration *prometheus.HistogramVec
	QueryTotal    *prometheus.CounterVec
	ErrorsTotal   *prometheus.CounterVec
}

// CustomMetrics пользовательские метрики
type CustomMetrics struct {
	//
}

// NewCollector создает и регистрирует все метрики
func NewCollector(appName string) *Collector {
	return &Collector{
		HTTP:     newHTTPMetrics(appName),
		Postgres: newPostgresMetrics(appName),
		Minio:    NewMinIOMetrics(appName),
		Redis:    newRedisMetrics(appName),
	}
}

func newHTTPMetrics(appName string) *HTTPMetrics {
	namespace := appName

	return &HTTPMetrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status", "handler"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),

		RequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 5), // 100, 1000, 10000, 100000, 1000000
			},
			[]string{"method", "path"},
		),

		ResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 5),
			},
			[]string{"method", "path", "status"},
		),

		RequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Current number of HTTP requests being served",
			},
		),
	}
}
