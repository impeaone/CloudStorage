package server

import (
	"CloudStorageProject-FileServer/internal/database/postgres"
	"CloudStorageProject-FileServer/internal/database/redis"
	"CloudStorageProject-FileServer/internal/metrics"
	"CloudStorageProject-FileServer/internal/middleware"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	consts "CloudStorageProject-FileServer/pkg/Constants"
	"CloudStorageProject-FileServer/pkg/config"
	"fmt"

	"context"
	"log/slog"
	"net/http"
	"sync"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	Port        string
	Logger      *slog.Logger
	Router      http.Handler
	Postgres    *postgres.Postgres
	Redis       *redis.Redis
	exitChan    chan struct{}
	connections *sync.WaitGroup
}

func NewServer(config *config.Config, logs *slog.Logger, pgs *postgres.Postgres, rds *redis.Redis,
	minio *minioClient.MinioClient, metric *metrics.HTTPMetrics) *Server {
	router := http.NewServeMux()
	// страницы
	// для static элементов (папка static)
	fs := http.FileServer(http.Dir(consts.StaticPath))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	//// site
	// перенаправление
	//router.HandleFunc("/", zeroPath)
	// страница входа
	router.HandleFunc("/index", indexPage)
	// страница с файлами
	router.HandleFunc("/client/api/v1/storage/", storagePage)
	// главная страница
	//TODO: ручку главной страницы

	// файловый api
	router.HandleFunc("GET /client/api/v1/get-file", getFileFunc)
	router.HandleFunc("POST /client/api/v1/upload-files", storeFilesFunc)
	router.HandleFunc("GET /client/api/v1/get-files-list", getFilesListFunc)
	router.HandleFunc("DELETE /client/api/v1/delete-file", deleteFilesFunc)

	//health check
	router.HandleFunc("/health", healthCheck)

	// Swagger: /swagger/index.html
	router.Handle("GET /swagger/", httpSwagger.WrapHandler)

	// Middleware
	exitChan := make(chan struct{})
	conns := new(sync.WaitGroup)
	ShutDown := middleware.ShutdownMiddleware(exitChan, conns, router)
	CheckPanics := middleware.PanicMiddleware(ShutDown, logs)
	HttpMetrics := metrics.HTTPMetricsMiddleware(CheckPanics, metric)
	validations := middleware.ValidateAPI(HttpMetrics, pgs, rds, minio, consts.TemplatePath, logs)
	handler := middleware.Logger(logs, validations)
	return &Server{
		Port:        config.ServerPort,
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

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf(":%s", s.Port), s.Router)
}
