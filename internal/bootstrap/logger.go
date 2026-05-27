package bootstrap

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() (*zap.Logger, error) {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	fmt.Print("current environment:", env)
	switch strings.ToLower(env) {
	case "local", "dev", "development", "test":
		config := zap.NewDevelopmentConfig()
		config.Encoding = "console"
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		return config.Build()

	case "staging", "stage":
		config := zap.NewProductionConfig()
		config.Encoding = "json"
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		return config.Build()

	case "production", "prod":
		config := zap.NewProductionConfig()
		config.Encoding = "json"
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		return config.Build()

	default:
		config := zap.NewProductionConfig()
		config.Encoding = "json"
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		return config.Build()
	}
}
