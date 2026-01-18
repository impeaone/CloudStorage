package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RedisMetrics метрики базы данных Redis
type RedisMetrics struct {
	QueryDuration *prometheus.HistogramVec
	QueryTotal    *prometheus.CounterVec
	ErrorsTotal   *prometheus.CounterVec
}

func newRedisMetrics(appName string) *RedisMetrics {
	namespace := appName

	return &RedisMetrics{
		QueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "redis_query_duration_seconds",
				Help:      "Duration of Redis queries in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		QueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_queries_total",
				Help:      "Total number of Redis queries",
			},
			[]string{"operation", "status"},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_errors_total",
				Help:      "Total number of Redis errors",
			},
			[]string{"error_type", "operation"},
		),
	}
}
