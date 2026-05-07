package bootstrap

import "go-echo-demo/internal/constants"

type AppConfig struct {
	ProjectName string
}

func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{
		ProjectName: constants.ProjectName,
	}
	return cfg, nil
}
