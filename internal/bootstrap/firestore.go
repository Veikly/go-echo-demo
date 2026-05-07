package bootstrap

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

// init firestore
func InitFireStore(ctx context.Context, projectName string) *firestore.Client {
	client, err := firestore.NewClient(ctx, projectName)
	if err != nil {
		fmt.Printf("Exception occurred while initializing the database connection %v\n", err)
		panic(err)
	}
	return client
}
