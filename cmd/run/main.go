package main

import (
	_ "CloudStorageProject-FileServer/docs"
	"CloudStorageProject-FileServer/internal/app"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"log"
	"log/slog"
	"os"
	"runtime"
)

/*
Требуемые переменные окружения:

	Runtime:
	tools.GetEnvAsInt("NUM_CPU", runtime.NumCPU())

	MINIO:
	tools.GetEnv("SERVER_PORT", "11682")
	tools.GetEnv("MINIO_ENDPOINT", "localhost:9000")
	tools.GetEnv("MINIO_EXAMPLE_BUCKET", "test")
	tools.GetEnv("MINIO_ROOT_USER", "user")
	tools.GetEnv("MINIO_ROOT_PASSWORD", "password")
	tools.GetEnvAsBool("MINIO_USER_SSL", false) это включает/выключает https

	Сам сервер:
	tools.GetEnvAsInt("SERVER_PORT", 11682)
	tools.GetEnv("SERVER_IP", "127.0.0.1")
	tools.GetEnv("SERVER_FILE_DIR", "C:/Files")

	Postgres
	tools.GetEnv("PG_USER", "postgres")
	tools.GetEnv("PG_PASSWORD", "080455mN")
	tools.GetEnv("PG_HOST", "localhost")
	tools.GetEnvAsInt("PG_PORT", 5432)
	tools.GetEnv("PG_DATABASE", "storage")
	tools.GetEnvAsBool("TEST_API_NEEDED", true);
	tools.GetEnv("TEST_API_KEY", "test")
	tools.GetEnv("TEST_API_EMAIL", "test@test.test")

	logger:
	tools.GetEnv("CloudStorage_LOGGER", "INFO")

	REDIS:
	tools.GetEnv("REDIS_HOST", "localhost")
	tools.GetEnvAsInt("REDIS_PORT", 6379)
	tools.GetEnv("REDIS_PASSWORD", "")
	rdsDB := tools.GetEnvAsInt("REDIS_DB", 0)

	Metrics:
	tools.GetEnvAsInt("MERICS_SERVER_PORT", 11680)
	tools.GetEnv("MERICS_SERVER_IP", "")
*/

// @title CloudStorage
// @version 1.0
// @description MinIO-base data storage
// @BasePath /client/api/v1
// @schemes http
func main() {
	ctx := context.Background()

	conf, err := config.Load(config.ConfPath)
	if err != nil {
		log.Fatal(err)
	}
	runtime.GOMAXPROCS(conf.NumCPU)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	ctx = context.WithValue(ctx, "logger", logger)
	ctx = context.WithValue(ctx, "config", conf)

	// Инициализация сервера
	application, err := app.NewApp(ctx)
	if err != nil {
		logger.Warn("application initialization error", "error", err, "place", tools.GetPlace())
		return
	}
	if err = application.Start(); err != nil {
		logger.Error("server start error", "error", err, "place", tools.GetPlace())
		os.Exit(1)
	}
	logger.Info("server stopped", "place", tools.GetPlace())
}
