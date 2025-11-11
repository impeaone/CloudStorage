package config

import (
	consts "CloudStorageProject-FileServer/pkg/Constants"
	"gopkg.in/yaml.v3"
	"os"
	"runtime"
)

type Config struct {
	Port      int    `yaml:"Port"`
	IPAddress string `yaml:"IPAddress"`
}

func ReadConfig() (*Config, error) {
	var ConfPath string
	if runtime.GOOS == "windows" {
		ConfPath = consts.ConfigPathWindows
	} else if runtime.GOOS == "linux" {
		ConfPath = consts.ConfigPathLinux
	} else {
		ConfPath = consts.ConfigPathLinux // Ну там, дааа, пока что так
	}
	bytes, err := os.ReadFile(ConfPath)
	if err != nil {
		return nil, err
	}
	var config Config
	errUnmarshal := yaml.Unmarshal(bytes, &config)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	return &config, nil
}
