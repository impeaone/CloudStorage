package app

import (
	"CloudStorageProject-FileServer/internal/app/server"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/logger/logger"
	"fmt"
	"net/http"
)

type App struct {
	fileServer *server.Server
}

func NewApp(config *config.Config, logger *logger.Log) *App {
	fileServer := server.NewServer(config, logger)
	return &App{
		fileServer: fileServer,
	}
}

func (app *App) Start() error {
	app.fileServer.Logger.Info(fmt.Sprintf("Server listening on port %d", app.fileServer.Port), logger.GetPlace())
	err := http.ListenAndServe(fmt.Sprintf(":%d", app.fileServer.Port), app.fileServer.Router)
	return err
}
