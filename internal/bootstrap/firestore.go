package bootstrap

import (
	"context"

	"cloud.google.com/go/firestore"
	"go.uber.org/zap"
)

// init firestore
func InitFireStore(ctx context.Context, projectName string) *firestore.Client {
	client, err := firestore.NewClient(ctx, projectName)
	if err != nil {
		zap.L().Error("Exception occurred while initializing the database connection", zap.Error(err))
		panic(err)
	}
	return client
}
