package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const ConfPath = ".env"

type Config struct {
	// CPU
	NumCPU int `env:"NUM_CPU" env-default:"4"`

	// MinIO
	MinIOEndpoint string `env:"MINIO_ENDPOINT" env-default:"minio:9000"`
	MinIOBucket   string `env:"MINIO_EXAMPLE_BUCKET" env-default:"test"`
	MinIOUser     string `env:"MINIO_ROOT_USER" env-default:"user"`
	MinIOPassword string `env:"MINIO_ROOT_PASSWORD" env-default:"password"`
	MinIOUseSSL   bool   `env:"MINIO_USER_SSL" env-default:"false"`

	// Server
	ServerPort string `env:"SERVER_PORT" env-default:"11682"`
	ServerIP   string `env:"SERVER_IP" env-default:"0.0.0.0"`

	// PostgreSQL
	PGUser     string `env:"PG_USER" env-default:"postgres"`
	PGPassword string `env:"PG_PASSWORD" env-default:"postgres"`
	PGHost     string `env:"PG_HOST" env-default:"postgres"`
	PGPort     string `env:"PG_PORT" env-default:"5432"`
	PGDatabase string `env:"PG_DATABASE" env-default:"storage"`

	// Test API
	TestAPINeeded bool   `env:"TEST_API_NEEDED" env-default:"true"`
	TestAPIKey    string `env:"TEST_API_KEY" env-default:"test"`
	TestAPIEmail  string `env:"TEST_API_EMAIL" env-default:"test@test.test"`

	// Redis
	RedisHost     string `env:"REDIS_HOST" env-default:"redis"`
	RedisPort     string `env:"REDIS_PORT" env-default:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" env-default:""`
	RedisDB       int    `env:"REDIS_DB" env-default:"0"`

	// Metrics
	MetricsServerPort string `env:"METRICS_SERVER_PORT" env-default:"11680"`
	MetricsServerIP   string `env:"METRICS_SERVER_IP" env-default:"0.0.0.0"`

	// Logging
	LogLevel string `env:"LOG_LEVEL" env-default:"INFO"`
}

func Load(envPath string) (*Config, error) {
	cfg := &Config{}

	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	if err := cfg.loadFromEnv(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadFromEnv() error {
	// CPU
	if val := os.Getenv("NUM_CPU"); val != "" {
		if num, err := strconv.Atoi(val); err == nil {
			c.NumCPU = num
		}
	}

	// MinIO
	if val := os.Getenv("MINIO_ENDPOINT"); val != "" {
		c.MinIOEndpoint = val
	}
	if val := os.Getenv("MINIO_EXAMPLE_BUCKET"); val != "" {
		c.MinIOBucket = val
	}
	if val := os.Getenv("MINIO_ROOT_USER"); val != "" {
		c.MinIOUser = val
	}
	if val := os.Getenv("MINIO_ROOT_PASSWORD"); val != "" {
		c.MinIOPassword = val
	}
	if val := os.Getenv("MINIO_USER_SSL"); val != "" {
		c.MinIOUseSSL = val == "true" || val == "1" || val == "yes"
	}

	// Server
	if val := os.Getenv("SERVER_PORT"); val != "" {
		c.ServerPort = val
	}
	if val := os.Getenv("SERVER_IP"); val != "" {
		c.ServerIP = val
	}

	// PostgreSQL
	if val := os.Getenv("PG_USER"); val != "" {
		c.PGUser = val
	}
	if val := os.Getenv("PG_PASSWORD"); val != "" {
		c.PGPassword = val
	}
	if val := os.Getenv("PG_HOST"); val != "" {
		c.PGHost = val
	}
	if val := os.Getenv("PG_PORT"); val != "" {
		c.PGPort = val
	}
	if val := os.Getenv("PG_DATABASE"); val != "" {
		c.PGDatabase = val
	}

	// Test API
	if val := os.Getenv("TEST_API_NEEDED"); val != "" {
		c.TestAPINeeded = val == "true" || val == "1" || val == "yes"
	}
	if val := os.Getenv("TEST_API_KEY"); val != "" {
		c.TestAPIKey = val
	}
	if val := os.Getenv("TEST_API_EMAIL"); val != "" {
		c.TestAPIEmail = val
	}

	// Redis
	if val := os.Getenv("REDIS_HOST"); val != "" {
		c.RedisHost = val
	}
	if val := os.Getenv("REDIS_PORT"); val != "" {
		c.RedisPort = val
	}
	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		c.RedisPassword = val
	}
	if val := os.Getenv("REDIS_DB"); val != "" {
		if db, err := strconv.Atoi(val); err == nil {
			c.RedisDB = db
		}
	}

	// Metrics
	if val := os.Getenv("METRICS_SERVER_PORT"); val != "" {
		c.MetricsServerPort = val
	}
	if val := os.Getenv("METRICS_SERVER_IP"); val != "" {
		c.MetricsServerIP = val
	}

	// Logging
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		c.LogLevel = val
	}

	return nil
}

func (c *Config) PostgreSQLDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.PGUser,
		c.PGPassword,
		c.PGHost,
		c.PGPort,
		c.PGDatabase,
	)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%s", c.ServerIP, c.ServerPort)
}

func (c *Config) MetricsAddr() string {
	return fmt.Sprintf("%s:%s", c.MetricsServerIP, c.MetricsServerPort)
}
