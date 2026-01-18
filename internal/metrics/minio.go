package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MinIOMetrics - метрики minio
type MinIOMetrics struct {
	UploadsTotal    *prometheus.CounterVec   // СКОЛЬКО ФАЙЛОВ ЗАГРУЖЕНО (всего)
	DownloadsTotal  *prometheus.CounterVec   // СКОЛЬКО ФАЙЛОВ СКАЧАНО (всего)
	FilesListTotal  *prometheus.CounterVec   // Количество файлов в бакете
	DeletesTotal    *prometheus.CounterVec   // Количество удаленных файлов
	UploadTime      *prometheus.HistogramVec // СКОЛЬКО ВРЕМЕНИ ЗАНИМАЕТ ЗАГРУЗКА
	UploadSize      *prometheus.HistogramVec // КАКОГО РАЗМЕРА ФАЙЛЫ ЗАГРУЖАЮТ
	UploadErrors    *prometheus.CounterVec   // ОШИБКИ ПРИ ЗАГРУЗКЕ
	DownloadErrors  *prometheus.CounterVec   // ОШИБКИ ПРИ СКАЧИВАНИИ
	FilesListErrors *prometheus.CounterVec   // ОШИБКИ ПРИ ПОЛУЧЕНИИ СПИСКА ФАЙЛОВ
	DeleteErrors    *prometheus.CounterVec   // Ошибки при удалении файлов
}

// NewMinIOMetrics - создает метрики
func NewMinIOMetrics(appName string) *MinIOMetrics {
	return &MinIOMetrics{
		UploadsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_uploads_total",
				Help:      "Всего загружено файлов",
			},
			[]string{"bucket"}, // группируем по bucket
		),

		FilesListTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_files_list_total",
				Help:      "Всего файлов в у пользователя в бакете",
			},
			[]string{"bucket"}, // группируем по bucket
		),

		DownloadsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_downloads_total",
				Help:      "Всего скачано файлов",
			},
			[]string{"bucket"},
		),

		DeletesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_deletes_total",
				Help:      "Всего удалено файлов",
			},
			[]string{"bucket"},
		),

		UploadTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: appName,
				Name:      "minio_upload_time_seconds",
				Help:      "Время загрузки файла",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"bucket"},
		),

		UploadSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: appName,
				Name:      "minio_upload_size_bytes",
				Help:      "Размер загружаемых файлов",
				Buckets: []float64{
					1024,      // 1KB
					10240,     // 10KB
					102400,    // 100KB
					1048576,   // 1MB
					10485760,  // 10MB
					104857600, // 100MB
				},
			},
			[]string{"bucket"},
		),

		UploadErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_upload_errors_total",
				Help:      "Ошибки при загрузке",
			},
			[]string{"bucket", "error"},
		),

		DeleteErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_delete_errors_total",
				Help:      "Ошибки при удалении файлов",
			},
			[]string{"bucket", "error"},
		),

		FilesListErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_files_list_errors_total",
				Help:      "Ошибки при выдаче списка файлов из  бакета пользователя",
			},
			[]string{"bucket", "error"},
		),

		DownloadErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: appName,
				Name:      "minio_download_errors_total",
				Help:      "Ошибки при скачивании",
			},
			[]string{"bucket", "error"},
		),
	}
}
