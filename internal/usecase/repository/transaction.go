package repository

import (
	"context"

	"cloud.google.com/go/firestore"
)

// *firestore.Client 自动持有此接口 也就是说 持有firestore.Client即实现了此方法 可以无缝调用
// 在初始化TransactionService时 传入全局的*firestore.Client即可
type TransactionService interface {
	// 仿照firestore.client.RunTransaction方法签名 写一个接口 让持有这个接口的service自动拥有
	RunTransaction(ctx context.Context, f func(context.Context, *firestore.Transaction) error, opts ...firestore.TransactionOption) error
}
