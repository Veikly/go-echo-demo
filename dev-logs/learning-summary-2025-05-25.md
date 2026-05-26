# Go + Echo + Clean Architecture 学习总结 (2025-05-25)

## 今日提交总览

| 时间 | 提交 | 知识点 |
|------|------|--------|
| 09:47 | 认证中间件集成 | Echo Middleware、Firebase Auth、全局变量管理 |
| 10:21 | 环境变量配置 | godotenv 加载 .env、配置结构体模式 |
| 10:48 | 任务枚举值常量化 | Go 自定义类型 + iota/常量枚举 |
| 11:42 | 统一化日志组件 | zap 全局 Logger、ReplaceGlobals、结构化日志 |
| 14:03 | 优雅停机 | os.Signal、goroutine、channel、Graceful Shutdown |
| 17:20 | 用户信息完善与查询接口 | 嵌套 struct、DTO 构造函数、全链路开发 |

---

## 1. Echo 中间件 & Firebase Auth 鉴权

### 核心模式：中间件函数签名

```go
func SomeMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // 前置处理（鉴权、日志、限流等）
        // ...
        return next(c)  // 放行到下一个中间件或 handler
    }
}
```

关键点：
- `echo.HandlerFunc` 即 `func(c echo.Context) error`，中间件本质是函数装饰器
- `c.Set("key", value)` 在 echo.Context 中存数据，后续 handler 通过 `c.Get("key")` 取出
- `echo.NewHTTPError(statusCode, message)` 返回 HTTP 错误，会触发 `HTTPErrorHandler`

### Firebase Auth 鉴权流程

```
Authorization: Bearer <idToken>
    → 解析 Bearer token
    → authClient.VerifyIDToken(ctx, idToken)
    → 验证通过：c.Set("firebase_token", token)，放行
    → 验证失败：返回 401
```

### 全局变量管理 AuthClient

```go
// bootstrap/firebase.go
var AuthClient *auth.Client

func InitFirebase() {
    // 初始化 Firebase App → 获取 Auth Client → 存入全局变量
}
```

- 全局变量 `AuthClient` 在 `main()` 启动时通过 `InitFirebase()` 初始化
- 中间件中直接引用 `bootstrap.AuthClient`，无需依赖注入传递
- 优点：简单直接；缺点：测试时不便 mock（现阶段够用）

---

## 2. 环境变量与配置管理

### godotenv 加载 .env 文件

```go
// bootstrap/config.go
func LoadConfig() (*AppConfig, error) {
    _ = godotenv.Load("../../.env")  // 加载 .env，文件不存在也不报错
    return &AppConfig{ProjectName: constants.ProjectName}, nil
}
```

- `godotenv.Load()` 将 .env 中的键值对加载为环境变量，可通过 `os.Getenv()` 读取
- `_ =` 忽略错误：.env 文件可以不存在，适用于生产环境用真正的环境变量

### 配置结构体模式

```go
type AppConfig struct {
    ProjectName string
}
```

- 所有配置集中在一个 struct 中，类型安全，IDE 有自动补全
- 后续可扩展到从 `os.Getenv()` 读端口、数据库地址等

---

## 3. 自定义枚举类型（类型安全替代裸 int）

### 改造前 vs 改造后

```go
// 改造前：裸 int，任意值都能传入
Status int  // 谁知道 1 是什么意思？

// 改造后：自定义类型 + 命名常量
type TaskStatus int

const (
    StatusTodo       TaskStatus = 0  // 待办
    StatusInProgress TaskStatus = 1  // 进行中
    StatusDone       TaskStatus = 2  // 已完成
    StatusAbandoned  TaskStatus = 3  // 已废弃
    StatusArchived   TaskStatus = 4  // 已归档
)
```

### 为什么这样做

1. **编译期类型检查**：函数签名 `func(ts TaskStatus)` 不接受普通 `int`，防止误传
2. **可读性**：`StatusDone` vs 数字 `2`
3. **IDE 友好**：自动补全列出所有可选值
4. **零成本**：底层还是 int，没有性能损耗

### 组织方式：`enums` 子包

```
internal/constants/enums/task.go
```

枚举放在 `constants/enums/` 下，与业务常量分离。

---

## 4. Zap 结构化日志 & 全局 Logger

### Zap 是什么

Uber 开源的高性能结构化日志库，比 Go 标准库 `log` 快 4-10 倍，支持 JSON 格式输出，方便 ELK/Grafana 采集。

### 初始化全局 Logger

```go
logger, _ := zap.NewProduction()
defer logger.Sync()          // 退出前 flush 缓冲区
zap.ReplaceGlobals(logger)   // 替换全局 logger
```

之后任何地方都可以直接使用：

```go
zap.L().Info("server starting", zap.String("port", ":8080"))
zap.L().Error("request failed", zap.Error(err))
zap.L().Fatal("init failed", zap.Error(err))
```

### 关键 API

| 方法 | 用途 |
|------|------|
| `zap.L()` | 获取全局 logger |
| `zap.NewProduction()` | 生产配置：JSON 格式、Info 级别 |
| `zap.NewDevelopment()` | 开发配置：人类可读格式、Debug 级别 |
| `zap.String(key, val)` | 结构化字段 |
| `zap.Error(err)` | 记录 error |
| `.Fatal()` | 记录 + `os.Exit(1)` |
| `.Sync()` | 刷新缓冲区（defer 调用） |

### 自定义请求日志中间件

```go
func ZapLogger(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        start := time.Now()
        err := next(c)  // 执行后续链
        zap.L().Info("HTTP Request Info",
            zap.String("method", c.Request().Method),
            zap.Int("status", c.Response().Status),
            zap.Float64("latency_seconds", time.Since(start).Seconds()),
        )
        return err
    }
}
```

核心技巧：`time.Now()` 在调用 `next()` 之前记录，之后计算差值得到请求耗时。

### 替换范围

| 原来 | 换成 |
|------|------|
| `log.Fatalf(...)` | `zap.L().Fatal(...)` |
| `fmt.Printf(...)` | `zap.L().Error(...)` |
| `gommon/log.Errorf(...)` | `zap.L().Error(...)` |
| `fmt.Println(...)` | `zap.L().Debug(...)` |
| `c.Logger().Errorf(...)` | `zap.L().Error(...)` |

---

## 5. 优雅停机（Graceful Shutdown）

### 为什么需要

K8s/Docker 发 SIGTERM 信号要你停时：
- 没有优雅停机 → 正在处理的请求直接断掉 → 用户看到 502
- 有优雅停机 → 等现有请求处理完（有超时上限）→ 再退出

### 实现模式

```go
// 1. 服务在 goroutine 中启动（不阻塞主线程）
go func() {
    if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
        zap.L().Fatal("server start error", zap.Error(err))
    }
}()

// 2. 主线程监听操作系统信号
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
sig := <-quit  // 阻塞，直到收到信号

// 3. 设置超时上下文，调用 Shutdown
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
e.Shutdown(shutdownCtx)  // 拒绝新请求，等待现有请求完成（最多 10s）
```

### 涉及的新知识点

| 概念 | 说明 |
|------|------|
| `goroutine` | `go func(){}()` 启动轻量级并发执行 |
| `channel` | `make(chan os.Signal, 1)` 创建缓冲区为 1 的信号通道 |
| `os.Signal` | 操作系统信号类型 |
| `syscall.SIGINT` | Ctrl+C 信号 |
| `syscall.SIGTERM` | K8s/Docker 发送的终止信号 |
| `signal.Notify` | 将指定信号转发到 channel |
| `context.WithTimeout` | 创建一个带超时的 context |
| `errors.Is(err, http.ErrServerClosed)` | 判断错误是否是"正常关闭" |

### 关键细节

- `e.Start(":8080")` 内部调用 `net/http` 的 `ListenAndServe`，正常情况下永不返回
- 当 `e.Shutdown()` 被调用后，`Start` 会返回 `http.ErrServerClosed`，这是**正常的关闭**不是错误
- `signal.Notify` 的第二个参数是可变参数，可以同时监听多个信号
- `quit` channel 缓冲区设为 1 防止信号丢失（SendSignal 是非阻塞的）

---

## 6. 用户信息完善与查询 —— 全链路实战

这是今天最大的 commit，实践了添加一个新实体的完整流程。

### 新增文件清单

```
模型层:   internal/model/user.go            → User, Address, Profile struct
枚举:     （无新增）
仓库接口: internal/usecase/repository/user.go → User 接口
用例 I/O: internal/usecase/usecaseio/user.go  → UserDetailOutput, CompleteUserInfoDetail
用例实现: internal/usecase/user.go            → GetMyDetail, CompleteUserInfo
Firestore:internal/infra/firestore/dto/user.go → User DTO + ToEntity/ToDTO
          internal/infra/firestore/service/user.go → Firestore 实现
请求 DTO: internal/request/user.go           → CompleteUserInfoInput
处理器:   internal/handler/user.go           → UserHandler
```

### 嵌套 struct 模式

```go
// model/user.go
type Address struct {
    Province string
    City     string
    Detail   string
}

type Profile struct {
    Avatar string
    Bio    string
}

type User struct {
    ID       string
    Username string
    Email    string
    Age      int
    Address  Address   // 嵌套 struct（值类型，不是指针）
    Profile  Profile
}
```

Firestore 自动将嵌套 struct 映射为嵌套 Map 字段：
```json
{
  "username": "zhangsan",
  "address": {
    "province": "Guangdong",
    "city": "Shenzhen"
  }
}
```

### DTO 转换模式 —— 构造函数法

这次 User 模块用了**构造函数模式**来做 DTO ↔ model 转换，比 Task 模块的逐字段赋值更干净：

```go
// usecaseio → model
func ToModelUser(info *CompleteUserInfoDetail) *model.User { ... }

// model → usecaseio
func NewUserDetailOutput(u *model.User) UserDetailOutput { ... }
func NewCompleteUserInfoDetail(u *model.User) CompleteUserInfoDetail { ... }
```

注意类型转换语法：
```go
model.Address(u.Address)   // 结构相同的 struct 可以直接类型转换
```

前提是两个 struct 的字段定义完全一致。

### 接口分拆

```go
// usecase/usecase.go
type TaskUseCase interface { ... }   // 任务相关
type UserUseCase interface { ... }   // 用户相关
```

不同实体的 usecase 接口分拆，符合接口隔离原则。

---

## 今日 Go 语言知识点速查表

| 知识点 | 出现位置 |
|--------|----------|
| 函数装饰器 / 闭包 | 中间件模式 |
| 自定义类型 `type X int` | TaskStatus 枚举 |
| `errors.Is(err, target)` | 错误类型判断（优雅停机） |
| `os.Signal` + `signal.Notify` | 信号监听 |
| goroutine + channel | 并发启动 server + 信号传递 |
| `context.WithTimeout` | 优雅停机超时控制 |
| `defer` | Sync 刷新、Cancel 释放 |
| 嵌套 struct | Address、Profile |
| struct 类型转换 `TypeA(b)` | DTO ↔ Entity 映射 |
| `zap.NewProduction` + `ReplaceGlobals` | 全局 logger |
| `c.Set` / `c.Get` | Echo Context 传值 |
| `firestore.MergeAll` | Firestore 部分字段更新 |
| 构造函数模式 `NewXxx()` | 全链路依赖注入 |

---

## 下一步建议

1. **测试** — 给 usecase 层写 table-driven 单元测试（mock repository 接口）
2. **slog** — 了解 Go 标准库 `log/slog`，理解它与 zap 的异同
3. **Context 深入** — `context.WithCancel`、`context.WithDeadline`、context 树
4. **错误处理进阶** — `%w` 包装、`errors.As`、自定义错误类型
5. **docker-compose** — 把 Firestore Emulator + 应用一起跑起来
