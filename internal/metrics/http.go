package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// responseWriter для перехвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

// HTTPMetricsMiddleware middleware для сбора HTTP метрик
func HTTPMetricsMiddleware(next http.Handler, c *HTTPMetrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Увеличиваем счетчик активных запросов
		c.RequestsInFlight.Inc()
		defer c.RequestsInFlight.Dec()

		// Перехватываем ResponseWriter
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Считаем размер запроса если возможно
		var requestSize float64
		if r.ContentLength > 0 {
			requestSize = float64(r.ContentLength)
		}
		// Выполняем обработку
		next.ServeHTTP(rw, r)

		// Рассчитываем метрики
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)

		// Метки для метрик
		labels := prometheus.Labels{
			"method":  r.Method,
			"path":    r.URL.Path,
			"status":  status,
			"handler": r.URL.Path, // или имя обработчика
		}

		// Обновляем метрики
		c.RequestsTotal.With(labels).Inc()
		c.RequestDuration.With(prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": status,
		}).Observe(duration)

		if requestSize > 0 {
			c.RequestSize.With(prometheus.Labels{
				"method": r.Method,
				"path":   r.URL.Path,
			}).Observe(requestSize)
		}

		c.ResponseSize.With(prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": status,
		}).Observe(float64(rw.written))
	})
}
