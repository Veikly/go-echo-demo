package provider

import (
	"go-echo-demo/internal/bootstrap"
	"go-echo-demo/internal/infra/firestore/transaction"
)

var GlobalTransationManger *transaction.TransactionManager

func InitTransactionManager() {
	GlobalTransationManger = transaction.NewTransactionManager(bootstrap.FirestoreClient)
}
