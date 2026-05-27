package bootstrap

import (
	"github.com/joho/godotenv"
)

type AppConfig struct {
	ProjectName string
}

// func LoadConfig() (*AppConfig, error) {
// 	// 加载环境变量
// 	_ = godotenv.Load("../../.env")
// 	cfg := &AppConfig{
// 		ProjectName: constants.ProjectName,
// 	}
// 	return cfg, nil
// }

func LoadConfig() {
	// 加载环境变量
	_ = godotenv.Load("../../.env")
}
