package MiniConfig

import (
	"CloudStorageProject-FileServer/pkg/config"
)

type MinioConfig struct {
	MinioExampleBucket string
	MinioEndPoint      string
	MinioRootUser      string
	MinioRootPassword  string
	MinioUserSSL       bool
}

func LoadMinioConfig(conf *config.Config) *MinioConfig {
	return &MinioConfig{
		MinioEndPoint:      conf.MinIOEndpoint,
		MinioExampleBucket: conf.MinIOBucket,
		MinioRootUser:      conf.MinIOUser,
		MinioRootPassword:  conf.MinIOPassword,
		MinioUserSSL:       conf.MinIOUseSSL,
	}
}
