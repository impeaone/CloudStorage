package MiniConfig

import (
	"CloudStorageProject-FileServer/pkg/tools"
)

type MinioConfig struct {
	Port               string
	MinioExampleBucket string
	MinioEndPoint      string
	MinioRootUser      string
	MinioRootPassword  string
	MinioUserSSL       bool
}

func LoadMinioConfig() *MinioConfig {
	port := tools.GetEnv("SERVER_PORT", "11682")
	minioEndPoint := tools.GetEnv("MINIO_ENDPOINT", "192.168.3.92:9000")
	minioExampleBucket := tools.GetEnv("MINIO_EXAMPLE_BUCKET", "test")
	minioRootUser := tools.GetEnv("MINIO_ROOT_USER", "user")
	minioRootPassword := tools.GetEnv("MINIO_ROOT_PASSWORD", "password")
	minioUserSSL := tools.GetEnvAsBool("MINIO_USER_SSL", false)

	return &MinioConfig{
		Port:               port,
		MinioEndPoint:      minioEndPoint,
		MinioExampleBucket: minioExampleBucket,
		MinioRootUser:      minioRootUser,
		MinioRootPassword:  minioRootPassword,
		MinioUserSSL:       minioUserSSL,
	}
}
