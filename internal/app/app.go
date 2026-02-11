package app

import (
	"CloudStorageProject-FileServer/internal/app/server"
	"CloudStorageProject-FileServer/internal/database/postgres"
	"CloudStorageProject-FileServer/internal/database/redis"
	"CloudStorageProject-FileServer/internal/metrics"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/logger/logger"
	"context"
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

func (app *App) ShutDown(ctx context.Context) error {
	if err := app.fileServer.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
