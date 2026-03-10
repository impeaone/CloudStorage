package middleware

import (
	"CloudStorageProject-FileServer/internal/database/postgres"
	"CloudStorageProject-FileServer/internal/database/redis"
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Logger - middleware, логги
func Logger(logs *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "static") && !strings.Contains(r.URL.String(), "swagger") {
			logs.Info("request url: "+r.URL.String(), "client", r.RemoteAddr, "method", r.Method,
				"time", time.Now().String(), "place", tools.GetPlace())
		}
		next.ServeHTTP(w, r)
	})
}

// ValidateAPI - middleware в котором валидируется api
func ValidateAPI(next http.Handler, pgs *postgres.Postgres, rds *redis.Redis,
	minio *minioClient.MinioClient, TmplPath string, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Валидация api
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		api := r.URL.Query().Get("api")
		if strings.Contains(r.URL.String(), "client") {
			//ключ проверяется тут
			if api == "" {
				logger.Warn("bad url api parameter", "client", r.RemoteAddr, "url", r.URL, "method", r.Method,
					"time", time.Now().String(), "place", tools.GetPlace())
				http.Error(w, "api is required", http.StatusBadRequest)
				return
			}

			if redisExists := rds.ExistsAPIField(api); !redisExists {
				existsPGS := pgs.CheckApiExists(api)
				if existsPGS == nil {
					logger.Warn("bad api", "client", r.RemoteAddr, "url", r.URL, "method", r.Method,
						"time", time.Now().String(), "place", tools.GetPlace())
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
						logger.Error("error to write api to redis", "error", err.Error(),
							"place", tools.GetPlace())
						return
					}
				}()
			}
			go func() {
				if err := pgs.UpdateLastLogin(api); err != nil {
					logger.Error("update last login postgres error", "error", err.Error(),
						"place", tools.GetPlace())
				}
				if err := rds.UpdateLastLogin(api); err != nil {
					logger.Error("update last login redis error", "error", err.Error(),
						"place", tools.GetPlace())
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

func PanicMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic middleware", "panic", err.(error).Error(), tools.GetPlace())
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
