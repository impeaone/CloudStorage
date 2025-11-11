package main

import (
	"CloudStorageProject-FileServer/internal/app"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/logger/logger"
	"fmt"
)

func main() {
	logs := logger.NewLog()

	conf, err := config.ReadConfig()
	if err != nil {
		logs.Error(fmt.Sprintf("Reading config file error: %v", err), logger.GetPlace())
		return
	}
	application := app.NewApp(conf, logs)
	if errStart := application.Start(); errStart != nil {
		logs.Error(fmt.Sprintf("Server Start error: %v", errStart), logger.GetPlace())
	}
}
