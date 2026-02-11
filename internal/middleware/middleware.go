package middleware

import (
	"CloudStorageProject-FileServer/internal/database/postgres"
	"CloudStorageProject-FileServer/internal/database/redis"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	logger2 "CloudStorageProject-FileServer/pkg/logger/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Logger - middleware, логги
func Logger(logs *logger2.Log, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "static") {
			logs.Info(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v",
				r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		}
		next.ServeHTTP(w, r)
	})
}

// ValidateAPI - middleware в котором валидируется api
func ValidateAPI(next http.Handler, pgs *postgres.Postgres, rds *redis.Redis,
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

			if redisExists := rds.ExistsAPIField(api); !redisExists {
				existsPGS := pgs.CheckApiExists(api)
				if existsPGS == nil {
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
				go func() {
					if err := rds.SetAPIField(existsPGS); err != nil {
						logger.Error("Ошибка добавления api записи в redis: "+err.Error(), logger2.GetPlace())
						return
					}
				}()
			}
			go func() {
				if err := pgs.UpdateLastLogin(api); err != nil {
					logger.Error("Ошибка обновления last_login  для api postgres: "+api, logger2.GetPlace())
				}
				if err := rds.UpdateLastLogin(api); err != nil {
					logger.Error("Ошибка обновления last_login  для api redis: "+api, logger2.GetPlace())
				}
			}()
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

func PanicMiddleware(next http.Handler, logger *logger2.Log) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic: "+err.(error).Error(), logger2.GetPlace())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ShutdownMiddleware - middleware проверяющий закрыт ли канал, для graceful shutdown
func ShutdownMiddleware(exitChan chan struct{}, conns *sync.WaitGroup, next http.Handler) http.Handler {
	conns.Add(1)
	defer conns.Done()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-exitChan:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "service_unavailable",
				"message": "Service is shutting down",
			})
			return
		default:
			next.ServeHTTP(w, r)
		}
	})
}
