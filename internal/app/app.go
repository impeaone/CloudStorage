package app

import (
	"CloudStorageProject-FileServer/internal/app/server"
	"CloudStorageProject-FileServer/internal/database/postgres"
	"CloudStorageProject-FileServer/internal/database/redis"
	"CloudStorageProject-FileServer/internal/metrics"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/closer"
	"CloudStorageProject-FileServer/pkg/config"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	fileServer   *server.Server
	metricServer *metrics.MetricsServer
	ctxCloser    *closer.Closer
	logger       *slog.Logger
	conf         *config.Config
}

func NewApp(ctx context.Context) (*App, error) {
	logger := ctx.Value("logger").(*slog.Logger)
	conf := ctx.Value("config").(*config.Config)

	ctxCloser := closer.NewCloser(logger)

	metric := metrics.NewCollector("CloudStorage")

	metricServer := metrics.NewMetricsServer(ctx)

	minio := minioClient.NewMinioClient(ctx, metric.Minio)
	if err := minio.Init(); err != nil {
		return nil, fmt.Errorf("minio init error: %w", err)
	}

	pgs, err := postgres.InitPostgres(ctx, metric.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres init error: %w", err)
	}

	rds, err := redis.NewRedis(ctx, metric.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis init error: %w", err)
	}

	fileServer := server.NewServer(conf, logger, pgs, rds, minio, metric.HTTP)

	ctxCloser.Add("minio", minio.CloseConnection)
	ctxCloser.Add("metrics", metricServer.Close)
	ctxCloser.Add("postgres", pgs.CloseConnection)
	ctxCloser.Add("redis", rds.CloseConnection)
	ctxCloser.Add("server", fileServer.Shutdown)
	return &App{
		fileServer:   fileServer,
		metricServer: metricServer,
		ctxCloser:    ctxCloser,
		logger:       logger,
		conf:         conf,
	}, nil
}

func (app *App) Start() error {
	if app == nil || app.fileServer == nil || app.metricServer == nil || app.ctxCloser == nil {
		return fmt.Errorf("application is nil")
	}

	errCh := make(chan error, 2)

	go func() {
		app.logger.Info("starting server", "port", app.fileServer.Port)
		errCh <- app.fileServer.Run()
	}()

	go func() {
		app.logger.Info("starting metrics", "port", app.conf.MetricsServerPort)
		errCh <- app.metricServer.StartMetricsServer()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigChan:
		app.logger.Info("signal received", "signal", sig.String())
	}
	return nil
}

func (app *App) Shutdown(ctx context.Context) error {
	return app.ctxCloser.Close(ctx)
}
