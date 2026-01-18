package postgres

import (
	"CloudStorageProject-FileServer/internal/metrics"
	"CloudStorageProject-FileServer/pkg/models"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool    *pgxpool.Pool
	ctxDB   context.Context
	clsDB   context.CancelFunc
	metrics *metrics.PostgresMetrics
}

func InitPostgres(metric *metrics.PostgresMetrics) (*Postgres, error) {
	pgUser := tools.GetEnv("PG_USER", "postgres")
	pgPassword := tools.GetEnv("PG_PASSWORD", "postgres")
	pgHost := tools.GetEnv("PG_HOST", "localhost")
	pgPort := tools.GetEnvAsInt("PG_PORT", 5432)
	pgDatabase := tools.GetEnv("PG_DATABASE", "storage")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", pgUser, pgPassword, pgHost, pgPort, pgDatabase)

	ctx := context.Background()
	pool, errPGX := pgxpool.New(ctx, connStr)
	if errPGX != nil {
		return nil, errPGX
	}
	err := createTables(pool, metric)
	if err != nil {
		return nil, err
	}
	createExampleAPI(pool)

	ctxDB, cls := context.WithCancel(context.Background())
	return &Postgres{
		pool:    pool,
		metrics: metric,
		ctxDB:   ctxDB,
		clsDB:   cls,
	}, nil

}
func createTables(pool *pgxpool.Pool, m *metrics.PostgresMetrics) error {
	ctx := context.Background()
	start := time.Now()
	_, err := pool.Query(ctx, `
		CREATE TABLE IF NOT EXISTS minio_keys (
    		id SERIAL PRIMARY KEY,
    		key_name VARCHAR(100) NOT NULL UNIQUE,
		    email VARCHAR(100) NOT NULL,
			cloud_access VARCHAR(5) DEFAULT '010',
    		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	duration := time.Since(start).Seconds()
	if err != nil {
		m.ErrorsTotal.WithLabelValues("query_error", "create_tables").Inc()
		m.QueryTotal.WithLabelValues("create_tables", "error").Inc()
		return err
	}
	m.QueryDuration.WithLabelValues("create_tables").Observe(duration)
	m.QueryTotal.WithLabelValues("create_tables", "success").Inc()
	return nil
}
func createExampleAPI(pool *pgxpool.Pool) {
	ctx := context.Background()
	if need := tools.GetEnvAsBool("TEST_API_NEEDED", true); !need {
		return
	}
	userTest := tools.GetEnv("TEST_API_KEY", "test")
	emailTest := tools.GetEnv("TEST_API_EMAIL", "test@test.test")

	query := fmt.Sprintf("INSERT INTO minio_keys (key_name, email) VALUES ('%s', '%s') on conflict do nothing", userTest, emailTest)
	_ = pool.QueryRow(ctx, query)
}

func (p *Postgres) CheckApiExists(api string) *models.APIPGS {
	start := time.Now()
	ctx := context.Background()
	apiStruct := &models.APIPGS{}
	query := fmt.Sprintf("SELECT * FROM minio_keys WHERE key_name = '%s'", api)
	if p.pool.QueryRow(ctx, query).Scan(&apiStruct.Id, &apiStruct.KeyName, &apiStruct.CloudAccess, &apiStruct.Email,
		&apiStruct.CreatedAt, &apiStruct.LastLogin); apiStruct.KeyName == "" {
		return nil
	}
	p.metrics.QueryTotal.WithLabelValues("check_api_exists", "success").Inc()
	p.metrics.QueryDuration.WithLabelValues("check_api_exists").Observe(time.Since(start).Seconds())
	return apiStruct
}
func (p *Postgres) UpdateLastLogin(api string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	_, err := p.pool.Exec(ctx, `UPDATE minio_keys SET last_login = $1 WHERE key_name = $2`, time.Now(), api)
	if err != nil {
		p.metrics.ErrorsTotal.WithLabelValues("query_error", "update_last_login").Inc()
		p.metrics.QueryTotal.WithLabelValues("update_last_login", "error").Inc()
		return fmt.Errorf("failed to update last login: %w", err)
	}
	p.metrics.QueryTotal.WithLabelValues("update_last_login", "success").Inc()
	p.metrics.QueryDuration.WithLabelValues("update_last_login").Observe(time.Since(start).Seconds())
	return nil
}

func (p *Postgres) Close() {
	p.clsDB()
}
