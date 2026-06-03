# Go + Echo + Clean Architecture 学习总结 (2026-06-02)

## 今日提交总览

| 时间 | 提交 | 知识点 |
|------|------|--------|
| 18:19 | feat(事务控制)：接入事务控制机制，在任务批量归档接口中使用 | TransactionService 接口设计、Firestore 事务三阶段模式、usecase 层控制事务边界 |

---

## 1. 业务场景设计：任务批量归档

### 为什么选这个场景

批量归档是一个天然需要事务控制的场景：

1. 多个 Task 文档的状态需要原子更新
2. 同时要维护 `user_stats` 文档中的 `archived_count` 计数器
3. 业务规则校验（只有「已完成」状态才能归档、只能操作自己的任务）必须在写入前完成

如果不用事务，可能出现：部分任务归档成功、部分失败，但计数器已经加了；或者计数器被并发写覆盖（丢失更新）。

### 新增的模型

```go
// internal/model/user_stats.go
type UserStats struct {
    UserID        string
    ArchivedCount int
    UpdatedAt     time.Time
}
```

`user_stats` 集合以 `userID` 作为文档 ID，每个用户一条记录，存储累计归档数。

---

## 2. 核心设计：TransactionService 接口

### 设计思路

事务控制的关键问题：如果把 `*firestore.Transaction` 封在 infra 层内部，不同 usecase 方法对同一个 repo 方法的事务要求不同，就没有办法灵活控制。

解决方案是让 **usecase 层决定是否开启事务**，infra 层只负责接收 `tx` 并在事务内执行操作。

```go
// internal/usecase/repository/transaction.go
type TransactionService interface {
    RunTransaction(ctx context.Context, f func(context.Context, *firestore.Transaction) error, opts ...firestore.TransactionOption) error
}
```

这个接口的签名和 `*firestore.Client.RunTransaction` **完全一致**，这意味着 `*firestore.Client` 天然实现了这个接口，装配时直接传入即可：

```go
// internal/bootstrap/provider/task.go
taskUseCase := usecase.NewTask(taskSvc, client)  // client 直接作为 TransactionService 传入
```

不需要写任何 adapter，零额外代码，这是 Go 隐式接口的优雅之处。

### 架构上的取舍

这个设计存在一个小的架构泄漏：`TransactionService` 依赖了 `cloud.google.com/go/firestore` 包（因为方法签名里有 `*firestore.Transaction`）。如果未来切换数据库，这个接口需要跟着改。

但在实际项目中这是可接受的范围——Firebase 已经作为基础设施固定下来，不会轻易替换，过度抽象反而增加维护成本。

---

## 3. 分层职责：事务边界在 usecase 层

### UseCase 层：决策层

usecase 方法知道"这个操作需要事务"，负责开启事务并把 `tx` 传给 repo：

```go
// internal/usecase/task.go
func (u *Task) BatchArchieveTask(ctx context.Context, ids []string) error {
    session, ok := domain.FromUserSession(ctx)
    if !ok {
        return constants.CredentialsAbsence
    }
    // usecase 决定开启事务
    return u.txSvc.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        return u.taskSvc.BatchArchieveTask(ctx, ids, session.UID, tx)
    })
}
```

注意接口签名里没有 `tx` 参数，handler 层完全不感知事务的存在：

```go
// internal/usecase/usecase.go
type TaskUseCase interface {
    // ...
    BatchArchieveTask(ctx context.Context, ids []string) error  // 干净，无 tx 参数
}
```

### Repository 接口：传递层

接口接收 `tx`，但自己不负责创建事务：

```go
// internal/usecase/repository/task.go
type TaskRepository interface {
    // ...
    BatchArchieveTask(ctx context.Context, ids []string, userID string, tx *firestore.Transaction) error
}
```

### Infra 层：执行层

infra 只管在事务内完成操作，三个阶段严格分开：

```go
// internal/infra/firestore/service/task.go
func (s *Task) BatchArchieveTask(ctx context.Context, ids []string, userID string, tx *firestore.Transaction) error {
    // === Phase 1: 事务内批量读取（所有 Get 必须在 Write 之前）===
    refs := make([]*firestore.DocumentRef, len(ids))
    for i, id := range ids {
        refs[i] = s.client.Collection("tasks").Doc(id)
    }
    snaps, err := tx.GetAll(refs)
    // ...
    statsSnap, err := tx.Get(statsRef)  // 读取计数器（允许 NotFound）

    // === Phase 2: 业务校验（基于快照，不做额外 IO）===
    for _, snap := range snaps {
        if task.CreatorID != userID { return constants.PermissionDenied }
        if task.Status != enums.StatusDone { return constants.TaskNotArchivable }
    }

    // === Phase 3: 所有写操作（在 Phase 1 读完之后才能执行）===
    for _, snap := range snaps {
        tx.Update(snap.Ref, []firestore.Update{
            {Path: "status", Value: enums.StatusArchived},
            {Path: "updated_at", Value: now},
        })
    }
    tx.Set(statsRef, dto.UserStats{
        ArchivedCount: currentCount + len(ids),  // read-modify-write
        // ...
    })
    return nil
}
```

---

## 4. Firestore 事务的核心规则

### 规则一：先读后写（Read-before-Write）

Firestore 事务内所有 `Get` / `GetAll` 必须在任何 `Set` / `Update` / `Delete` 之前执行，这是硬性约束，违反会报错。

```go
// ✅ 正确：先读完，再写
snaps, _ := tx.GetAll(refs)       // 读
statsSnap, _ := tx.Get(statsRef)  // 读
// ... 校验 ...
tx.Update(snap.Ref, updates)      // 写
tx.Set(statsRef, newStats)        // 写

// ❌ 错误：读写交替
tx.Update(ref1, updates)  // 写
snap, _ := tx.Get(ref2)   // 这里会报错
```

### 规则二：事务自动重试

`RunTransaction` 遇到并发写冲突（其他事务修改了你读取的文档）会自动重试，默认最多重试 5 次。这就是为什么回调函数必须是**幂等的**，不能在里面做发邮件、扣余额这类有副作用的操作。

### 规则三：Transaction vs WriteBatch 的选择

| 场景 | 选择 |
|------|------|
| 写之前需要读取文档并基于读取结果决定写什么 | **Transaction** |
| 只需要原子执行多个写操作，不依赖读取结果 | **WriteBatch**（更高效，无重试开销） |

本场景需要读取每个 task 的状态做校验，并且读取 `user_stats` 做 read-modify-write，所以必须用 Transaction。

### 规则四：read-modify-write 模式

更新计数器不能用 `FieldTransformIncrement`（事务内有限制），需要在事务内先读出当前值，再写入新值：

```go
// 在事务内读到 currentCount 后
tx.Set(statsRef, dto.UserStats{
    ArchivedCount: currentCount + len(ids),  // 基于读到的值做增量
})
```

事务保证了从读到写的原子性，不会有两个并发请求都读到 `currentCount=5`，然后都写入 `6` 导致丢失更新。

---

## 5. 发现并修复的设计问题

在实现过程中发现了框架搭建时的几个问题：

### 问题一：事务参数不应该暴露给接口

```go
// ❌ 错误：handler 层能看到 tx 参数，违反分层原则
BatchArchieveTask(ctx context.Context, ids []string, tx repository.TransactionService) error

// ✅ 正确：handler 不感知事务
BatchArchieveTask(ctx context.Context, ids []string) error
```

### 问题二：接口实现签名不一致

```go
// 接口定义（值类型）
BatchArchieveTask(ctx context.Context, ids []string, tx repository.TransactionService) error

// 实现（指针类型）——Go 里这是两个不同的类型，编译失败
func (u *Task) BatchArchieveTask(ctx context.Context, ids []string, tx *repository.TransactionService) error
```

### 问题三：RunTransaction 返回值被丢弃

```go
// ❌ 错误：事务的 error 被忽略，方法永远返回 nil
u.txSvc.RunTransaction(ctx, func(...) error { ... })
return nil

// ✅ 正确：直接透传 error
return u.txSvc.RunTransaction(ctx, func(...) error { ... })
```

---

## 6. 完整的请求链路

```
POST /api/tasks/archive
    Body: { "ids": ["id1", "id2", "id3"] }

→ handler.BatchArchieveTask()
    Bind + Validate（限制 1-20 个 ID）

→ usecase.BatchArchieveTask(ctx, ids)
    session 校验
    txSvc.RunTransaction() 开启事务

→ (事务内) taskSvc.BatchArchieveTask(ctx, ids, userID, tx)
    Phase 1: tx.GetAll(taskRefs) + tx.Get(statsRef)
    Phase 2: 权限校验 + 状态校验
    Phase 3: tx.Update(所有 task) + tx.Set(statsRef)

← 事务提交成功 → 200 OK
← 任意校验失败 → 事务回滚 → 对应业务错误码
```

---

## 今日 Go 语言知识点速查表

| 知识点 | 出现位置 |
|--------|----------|
| 隐式接口：`*firestore.Client` 自动实现 `TransactionService` | `repository/transaction.go` |
| 事务边界由 usecase 层控制，infra 层只负责执行 | `usecase/task.go` |
| Firestore 事务三阶段：先读全部 → 业务校验 → 批量写 | `service/task.go` |
| `tx.GetAll()` 批量读取多个文档 | `service/task.go` |
| read-modify-write 模式保证计数器并发安全 | `user_stats` 计数器更新 |
| 事务回调必须幂等（因为会自动重试） | `RunTransaction` 使用规范 |
| Transaction vs WriteBatch 的选择依据 | 本场景需要读后写，选 Transaction |
| 接口方法签名不应暴露基础设施细节（tx）给上层 | `usecase.go` 接口设计 |

---

## 下一步建议

1. **`BatchArchieveTask` 限流** — 当前通过 `validate:"min=1,max=20"` 限制了最多 20 个 ID，可以考虑在 usecase 层也加一道防御
2. **事务测试** — Firestore 事务无法在纯单元测试中验证，可以用 [Firestore Emulator](https://firebase.google.com/docs/emulator-suite) 做集成测试
3. **WriteBatch 对比实践** — 找一个不需要读取数据就能执行的批量写场景（比如批量删除已归档任务），用 `WriteBatch` 实现，体会两者的区别
4. **`user_stats` 初始化** — 目前 `NotFound` 时 `currentCount` 默认为 0，可以考虑在用户注册时就创建这条记录，避免每次归档都走 NotFound 分支
