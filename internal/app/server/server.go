package server

import (
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/logger/logger"
	"fmt"
	"net/http"
	"time"
)

var FilesDirectory string

type Server struct {
	Port   int
	Logger *logger.Log
	Router http.Handler
}

// Logger - middleware
func Logger(logs *logger.Log, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logs.Info(fmt.Sprintf("Client: %s; URL: %s; Method: %s; Time: %v",
			r.RemoteAddr, r.Host, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger.GetPlace())

		next.ServeHTTP(w, r)
	})
}

func NewServer(config *config.Config, logs *logger.Log) *Server {
	port := config.Port
	FilesDirectory = config.FilesDir
	//TODO: сделать проверку на наличие директории
	router := http.NewServeMux()
	router.HandleFunc("/client/get-files", getFilesFunc)

	handler := Logger(logs, router)
	return &Server{
		Port:   port,
		Logger: logs,
		Router: handler,
	}
}
