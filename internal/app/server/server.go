package server

import (
	"CloudStorageProject-FileServer/internal/database/postgres"
	"CloudStorageProject-FileServer/internal/database/redis"
	"CloudStorageProject-FileServer/internal/metrics"
	"CloudStorageProject-FileServer/internal/middleware"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/Constants"
	"CloudStorageProject-FileServer/pkg/config"
	logger2 "CloudStorageProject-FileServer/pkg/logger/logger"
	"context"
	"net/http"
	"runtime"
	"sync"
)

type Server struct {
	Port        int
	Logger      *logger2.Log
	Router      http.Handler
	Postgres    *postgres.Postgres
	Redis       *redis.Redis
	exitChan    chan struct{}
	connections *sync.WaitGroup
}

func NewServer(config *config.Config, logs *logger2.Log, pgs *postgres.Postgres, rds *redis.Redis,
	minio *minioClient.MinioClient, metric *metrics.HTTPMetrics) *Server {
	port := config.Port
	StaticPath := ""
	TemplatePath := ""
	if runtime.GOOS == "windows" {
		StaticPath = Constants.StaticPathWindows
		TemplatePath = Constants.TemplatePathWindows
	} else if runtime.GOOS == "darwin" {
		StaticPath = Constants.StaticPathDarwin
		TemplatePath = Constants.TemplatePathDarwin
	} else {
		StaticPath = Constants.StaticPathLinux
		TemplatePath = Constants.TemplatePathLinux
	}

	//TODO: Minio делаем
	router := http.NewServeMux()

	// страницы
	// для static элементов (папка static)
	fs := http.FileServer(http.Dir(StaticPath))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	//// site
	// перенаправление
	router.HandleFunc("/", zeroPath)
	// страница входа
	router.HandleFunc("/index", indexPage)
	// страница с файлами
	router.HandleFunc("/client/api/v1/storage/", storagePage)
	// главная страница
	//TODO: ручку главной страницы

	// файловый api
	router.HandleFunc("/client/api/v1/get-file", getFileFunc)
	router.HandleFunc("/client/api/v1/upload-files", storeFilesFunc)
	router.HandleFunc("/client/api/v1/get-files-list", getFilesListFunc)
	router.HandleFunc("/client/api/v1/delete-file", deleteFilesFunc)

	//health check
	router.HandleFunc("/health", healthCheck)

	// Middleware
	exitChan := make(chan struct{})
	conns := new(sync.WaitGroup)
	ShutDown := middleware.ShutdownMiddleware(exitChan, conns, router)
	CheckPanics := middleware.PanicMiddleware(ShutDown, logs)
	HttpMetrics := metrics.HTTPMetricsMiddleware(CheckPanics, metric)
	validations := middleware.ValidateAPI(HttpMetrics, pgs, rds, minio, TemplatePath, logs)
	handler := middleware.Logger(logs, validations)
	return &Server{
		Port:        port,
		Logger:      logs,
		Router:      handler,
		exitChan:    exitChan,
		Postgres:    pgs,
		Redis:       rds,
		connections: conns,
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	close(s.exitChan)

	finished := make(chan struct{})
	go func() {
		s.connections.Wait()
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
