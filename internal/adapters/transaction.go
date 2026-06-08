package adapters

import "context"

// 事务的抽象标识 用于在usecase层表示有事务 但不表明具体实现
type TxContent interface{}

// 业务函数签名 需要进行事务控制的业务逻辑 作为回调函数传入
type TxFunc func(ctx context.Context) error

// 抽象事务管理器 基础设施层提供具体实现
type TransactionManager interface {
	// 运行事务
	RunInTransaction(ctx context.Context, tx TxFunc) error
}

type txKey struct{}

// 将事务塞入上下文
func ContextWithTx(ctx context.Context, tx TxContent) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// 从上下文中获取事务信息 如果没有则代表无需事务控制
func TxFromContext(ctx context.Context) (TxContent, bool) {
	tx, ok := ctx.Value(txKey{}).(TxContent)
	return tx, ok
}
