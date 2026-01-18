package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func newPostgresMetrics(appName string) *PostgresMetrics {
	namespace := appName

	return &PostgresMetrics{
		QueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "postgres_query_duration_seconds",
				Help:      "Duration of PostgreSQL queries in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		QueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "postgres_queries_total",
				Help:      "Total number of PostgreSQL queries",
			},
			[]string{"operation", "status"},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "postgres_errors_total",
				Help:      "Total number of PostgreSQL errors",
			},
			[]string{"error_type", "operation"},
		),
	}
}
