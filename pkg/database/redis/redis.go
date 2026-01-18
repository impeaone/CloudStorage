package redis

import (
	"CloudStorageProject-FileServer/internal/metrics"
	"CloudStorageProject-FileServer/pkg/models"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	pool    *redis.Client
	metrics *metrics.RedisMetrics
}

func NewRedis(metric *metrics.RedisMetrics) (*Redis, error) {
	rdsHost := tools.GetEnv("REDIS_HOST", "localhost")
	rdsPort := tools.GetEnvAsInt("REDIS_PORT", 6379)
	rdsPassword := tools.GetEnv("REDIS_PASSWORD", "")
	rdsDB := tools.GetEnvAsInt("REDIS_DB", 0)
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", rdsHost, rdsPort),
		Password: rdsPassword,
		DB:       rdsDB,
	})
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		metric.ErrorsTotal.WithLabelValues("redis_ping").Inc()
		return nil, err
	}
	return &Redis{
		pool:    client,
		metrics: metric,
	}, nil
}

func (rds *Redis) Close() {
	rds.pool.Close()
}

func (rds *Redis) SetAPIField(apiData *models.APIPGS) error {
	ctx := context.Background()
	start := time.Now()
	_, err := rds.pool.HSet(ctx, "apikey:"+apiData.KeyName, map[string]interface{}{
		"id":          apiData.Id,
		"name":        apiData.KeyName,
		"email":       apiData.Email,
		"createdAt":   apiData.CreatedAt,
		"lastLogin":   apiData.LastLogin,
		"cloudAccess": apiData.CloudAccess,
	}).Result()
	if err != nil {
		rds.metrics.ErrorsTotal.WithLabelValues("redis_set_api").Inc()
		rds.metrics.QueryTotal.WithLabelValues("redis_set_api", "error").Inc()
		return err
	}
	rds.metrics.QueryTotal.WithLabelValues("redis_set_api", "success").Inc()
	rds.metrics.QueryDuration.WithLabelValues("redis_set_api").Observe(time.Since(start).Seconds())
	return nil
}

func (rds *Redis) GetAPIField(apikey string) (*models.APIPGS, error) {
	ctx := context.Background()
	start := time.Now()
	user, err := rds.pool.HGetAll(ctx, fmt.Sprintf("apikey:%s", apikey)).Result()
	if err != nil {
		rds.metrics.ErrorsTotal.WithLabelValues("redis_get_api").Inc()
		rds.metrics.QueryTotal.WithLabelValues("redis_get_api", "error").Inc()
		return nil, err
	}
	id, _ := strconv.Atoi(user["id"])
	cloudAccess := user["cloudAccess"]
	email := user["email"]
	CreatedAt, _ := time.Parse("2006-01-02 15:04:05", user["createdAt"])
	LastLogin, _ := time.Parse("2006-01-02 15:04:05", user["lastLogin"])

	rds.metrics.QueryTotal.WithLabelValues("redis_get_api", "success").Inc()
	rds.metrics.QueryDuration.WithLabelValues("redis_get_api").Observe(time.Since(start).Seconds())
	return &models.APIPGS{
		Id:          id,
		KeyName:     apikey,
		Email:       email,
		CloudAccess: cloudAccess,
		CreatedAt:   CreatedAt,
		LastLogin:   LastLogin,
	}, nil
}

func (rds *Redis) DelAPIField(apikey string) {
	ctx := context.Background()
	rds.pool.HDel(ctx, fmt.Sprintf("apikey:%s", apikey))
}

func (rds *Redis) ExistsAPIField(apikey string) bool {
	ctx := context.Background()
	start := time.Now()
	exist, err := rds.pool.Exists(ctx, fmt.Sprintf("apikey:%s", apikey)).Result()
	if err != nil {
		rds.metrics.ErrorsTotal.WithLabelValues("redis_exists_api").Inc()
		rds.metrics.QueryTotal.WithLabelValues("redis_exists_api", "error").Inc()
		return false
	}

	rds.metrics.QueryTotal.WithLabelValues("redis_api_exists", "success").Inc()
	rds.metrics.QueryDuration.WithLabelValues("redis_api_exists").Observe(time.Since(start).Seconds())
	if exist == 0 {
		return false
	}
	return true
}

func (rds *Redis) UpdateLastLogin(apikey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	// делаем транзакцию
	pipeline := rds.pool.Pipeline()

	pipeline.HSet(ctx, "apikey:"+apikey, "lastLogin", time.Now())
	pipeline.Expire(ctx, "apikey:"+apikey, 14*24*time.Hour)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		rds.metrics.ErrorsTotal.WithLabelValues("redis_update_last_login").Inc()
		rds.metrics.QueryTotal.WithLabelValues("redis_update_last_login", "error").Inc()
		return err
	}
	rds.metrics.QueryTotal.WithLabelValues("redis_update_last_login", "success").Inc()
	rds.metrics.QueryDuration.WithLabelValues("redis_update_last_login").Observe(time.Since(start).Seconds())
	return nil
}
