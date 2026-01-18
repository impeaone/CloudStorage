package main

import (
	"CloudStorageProject-FileServer/internal/app"
	"CloudStorageProject-FileServer/internal/metrics"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/database/postgres"
	"CloudStorageProject-FileServer/pkg/database/redis"
	"CloudStorageProject-FileServer/pkg/logger/logger"
	"CloudStorageProject-FileServer/pkg/tools"
	"fmt"
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
func main() {
	runtime.GOMAXPROCS(tools.GetEnvAsInt("NUM_CPU", runtime.NumCPU()))

	mtrcs := metrics.NewCollector("CloudStorage")

	// Logger
	logs := logger.NewLog(tools.GetEnv("CloudStorage_LOGGER", "INFO"))

	// Инициализация Minio
	minio := minioClient.NewMinioClient(mtrcs.Minio)
	if errMinio := minio.Init(); errMinio != nil {
		logs.Error(fmt.Sprintf("Ошибка инициализации Minio: %v", errMinio), logger.GetPlace())
		return
	}
	logs.Info("Успешная инициализация Minio", logger.GetPlace())
	// Инициализаяция Postgres
	pgs, errPGS := postgres.InitPostgres(mtrcs.Postgres)
	if errPGS != nil {
		logs.Error(fmt.Sprintf("Ошибка инициализации PostgreSQL: %v", errPGS), logger.GetPlace())
		return
	}
	logs.Info("Успешная инициализация PostgreSQL", logger.GetPlace())

	rds, errRds := redis.NewRedis(mtrcs.Redis)
	if errRds != nil {
		logs.Error(fmt.Sprintf("Ошибка инициализации Redis: %v", errRds), logger.GetPlace())
		return
	}
	logs.Info("Успешная инициализация Redis", logger.GetPlace())
	// Инициализация конфига
	conf, err := config.ReadConfig()
	if err != nil {
		logs.Error(fmt.Sprintf("Reading config file error: %v", err), logger.GetPlace())
		return
	}
	logs.Info("Успешная инициализация конфига", logger.GetPlace())

	// Запуск подсервера с метриками
	metrics.StartMetricsServer(logs, mtrcs)

	// Инициализация сервера
	application := app.NewApp(conf, logs, pgs, rds, minio, mtrcs.HTTP)
	if errStart := application.Start(); errStart != nil {
		logs.Error(fmt.Sprintf("Server Start error: %v", errStart), logger.GetPlace())
		return
	}
}
