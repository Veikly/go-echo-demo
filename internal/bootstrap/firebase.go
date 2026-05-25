package bootstrap

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

func InitFirebase() {
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "/Users/lin/go-study/go-echo-demo/go-echo-demo-firebase-adminsdk-fbsvc-98c4d740c5.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		zap.L().Fatal("Firebase 初始化失败", zap.Error(err))
	}
	AuthClient, err = app.Auth(context.Background())
	if err != nil {
		zap.L().Fatal("获取AuthClient失败", zap.Error(err))
	}
}
