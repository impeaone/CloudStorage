package app

import (
	"CloudStorageProject-FileServer/internal/app/server"
	"CloudStorageProject-FileServer/internal/metrics"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/database/postgres"
	"CloudStorageProject-FileServer/pkg/database/redis"
	"CloudStorageProject-FileServer/pkg/logger/logger"
	"fmt"
	"net/http"
)

type App struct {
	fileServer *server.Server
}

func NewApp(config *config.Config, logger *logger.Log, pgs *postgres.Postgres, rds *redis.Redis,
	minio *minioClient.MinioClient, metric *metrics.HTTPMetrics) *App {
	fileServer := server.NewServer(config, logger, pgs, rds, minio, metric)
	return &App{
		fileServer: fileServer,
	}
}

func (app *App) Start() error {
	app.fileServer.Logger.Info(fmt.Sprintf("Server listening on port %d", app.fileServer.Port), logger.GetPlace())
	err := http.ListenAndServe(fmt.Sprintf(":%d", app.fileServer.Port), app.fileServer.Router)
	return err
}
