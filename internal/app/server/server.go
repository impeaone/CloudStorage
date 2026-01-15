package server

import (
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/Constants"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/database/postgres"
	"CloudStorageProject-FileServer/pkg/database/redis"
	logger2 "CloudStorageProject-FileServer/pkg/logger/logger"
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Server struct {
	Port     int
	Logger   *logger2.Log
	Router   http.Handler
	Postgres *postgres.Postgres
	Redis    *redis.Redis
}

// Logger - middleware
func Logger(logs *logger2.Log, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "static") {
			logs.Info(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v",
				r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		}
		next.ServeHTTP(w, r)
	})
}
func validate(next http.Handler, pgs *postgres.Postgres, rds *redis.Redis,
	minio *minioClient.MinioClient, TmplPath string, logger *logger2.Log) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Валидация api
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		api := r.URL.Query().Get("api")
		if strings.Contains(r.URL.String(), "client") {
			//ключ проверяется тут
			if api == "" {
				logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: bad url api parameter",
					r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
				http.Error(w, "api is required", http.StatusBadRequest)
				return
			}

			// TODO: потом тут redis еще

			if exists := pgs.CheckApiExists(api); !exists {
				logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: bad api",
					r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
				http.SetCookie(w, &http.Cookie{
					Name:    "apikey",
					Value:   "",
					Path:    "/",
					MaxAge:  -1,
					Expires: time.Unix(0, 0),
				})
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		r = r.WithContext(context.WithValue(r.Context(), "api", api))
		r = r.WithContext(context.WithValue(r.Context(), "postgres", pgs))
		r = r.WithContext(context.WithValue(r.Context(), "redis", rds))
		r = r.WithContext(context.WithValue(r.Context(), "minio", minio))
		r = r.WithContext(context.WithValue(r.Context(), "tmplPath", TmplPath))
		r = r.WithContext(context.WithValue(r.Context(), "logger", logger))
		next.ServeHTTP(w, r)
	})
}

func NewServer(config *config.Config, logs *logger2.Log, pgs *postgres.Postgres, rds *redis.Redis,
	minio *minioClient.MinioClient) *Server {
	port := config.Port
	StaticPath := ""
	TemplatePath := ""
	if runtime.GOOS == "windows" {
		StaticPath = Constants.StaticPathWindows
		TemplatePath = Constants.TemplatePathWindows
	} else if runtime.GOOS == "linux" {
		StaticPath = Constants.StaticPathLinux
		TemplatePath = Constants.TemplatePathLinux
	} else {
		StaticPath = Constants.StaticPathLinux
		TemplatePath = Constants.TemplatePathLinux
	}

	//TODO: Minio делаем
	router := http.NewServeMux()

	// страницы
	// для static элементов (папка static)
	fs := http.FileServer(http.Dir(StaticPath))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	//// site
	// перенаправление
	router.HandleFunc("/", zeroPath)
	// страница входа
	router.HandleFunc("/index", indexPage)
	// страница с файлами
	router.HandleFunc("/client/api/v1/storage/", storagePage)
	// главная страница
	//TODO: ручку главной страницы

	// файловый api
	router.HandleFunc("/client/api/v1/get-file", getFileFunc)
	router.HandleFunc("/client/api/v1/upload-files", storeFilesFunc)
	router.HandleFunc("/client/api/v1/get-files-list", getFilesListFunc)
	router.HandleFunc("/client/api/v1/delete-file", deleteFilesFunc)

	//health check
	router.HandleFunc("/health", healthCheck)

	//// дополнительно
	// для главной страницы
	// ручка предоставляющая список точек на карте яндекс
	// TODO: сделать эту ручку
	// ручка предоставляющая список новостей
	// TODO: сделать эту ручку

	// Сдлеать ручку для админки
	// TODO: сделать

	validations := validate(router, pgs, rds, minio, TemplatePath, logs)
	handler := Logger(logs, validations)
	return &Server{
		Port:   port,
		Logger: logs,
		Router: handler,
	}
}
