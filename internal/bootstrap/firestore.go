package bootstrap

import (
	"context"

	"cloud.google.com/go/firestore"
)

// init firestore
func InitFireStore(ctx context.Context, projectName string) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, projectName)
	if err != nil {
		return nil, err
	}
	return client, nil
}
