package main

import (
	"CloudStorageProject-FileServer/internal/app"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/database/postgres"
	"CloudStorageProject-FileServer/pkg/database/redis"
	"CloudStorageProject-FileServer/pkg/logger/logger"
	"CloudStorageProject-FileServer/pkg/tools"
	"fmt"
)

/*
Требуемые переменные окружения:

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

	logger:
	tools.GetEnv("CloudStorage_LOGGER", "INFO")
*/
func main() {
	// Logger
	logs := logger.NewLog(tools.GetEnv("CloudStorage_LOGGER", "INFO"))

	// Инициализация Minio
	minio := minioClient.NewMinioClient()
	if errMinio := minio.Init(); errMinio != nil {
		logs.Error(fmt.Sprintf("Ошибка инициализации Minio: %v", errMinio), logger.GetPlace())
		return
	}
	logs.Info("Успешная инициализация Minio", logger.GetPlace())
	// Инициализаяция Postgres
	pgs, errPGS := postgres.InitPostgres()
	if errPGS != nil {
		logs.Error(fmt.Sprintf("Ошибка инициализации PostgreSQL: %v", errPGS), logger.GetPlace())
		return
	}
	logs.Info("Успешная инициализация PostgreSQL", logger.GetPlace())

	rds := &redis.Redis{}

	// Инициализация конфига
	conf, err := config.ReadConfig()
	if err != nil {
		logs.Error(fmt.Sprintf("Reading config file error: %v", err), logger.GetPlace())
		return
	}
	logs.Info("Успешная инициализая конфига", logger.GetPlace())
	// Инициализация сервера
	application := app.NewApp(conf, logs, pgs, rds, minio)
	if errStart := application.Start(); errStart != nil {
		logs.Error(fmt.Sprintf("Server Start error: %v", errStart), logger.GetPlace())
		return
	}
}
