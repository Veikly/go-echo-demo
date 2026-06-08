package transaction

import (
	"context"
	"go-echo-demo/internal/adapters"

	"cloud.google.com/go/firestore"
)

// 对*firestore.Transaction做一层薄封装
type FirestoreTx struct {
	Tx *firestore.Transaction
}

// Firestore事务管理器
type TransactionManager struct {
	client *firestore.Client
}

func NewTransactionManager(client *firestore.Client) *TransactionManager {
	return &TransactionManager{
		client: client,
	}
}

// 管理器实现抽象事务管理器
func (m *TransactionManager) RunInTransaction(ctx context.Context, fn adapters.TxFunc) error {
	return m.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// 将事务开关包装进上下文
		txCtx := adapters.ContextWithTx(ctx, &FirestoreTx{Tx: tx})
		return fn(txCtx)
	})
}
