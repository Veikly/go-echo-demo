package bootstrap

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client
var FirestoreClient *firestore.Client

func InitFirebase(ctx context.Context) {
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "/Users/lin/go-study/go-echo-demo/go-echo-demo-firebase-adminsdk-fbsvc-98c4d740c5.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		zap.L().Fatal("FirebaseApp 初始化失败", zap.Error(err))
	}
	AuthClient, err = app.Auth(context.Background())
	if err != nil {
		zap.L().Fatal("获取AuthClient失败", zap.Error(err))
	}
	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		zap.L().Fatal("初始化FirestoreClient失败", zap.Error(err))
	}
}
