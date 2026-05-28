# Go + Echo + Clean Architecture 学习总结 (2025-05-27)

## 今日提交总览

| 时间 | 提交 | 知识点 |
|------|------|--------|
| 10:57 | 简化 Firebase 初始化逻辑 | 全局变量管理、Firebase Admin SDK、bootstrap 启动流程 |
| 13:41 | 用户管理接口逻辑完善 | 首次登录自动创建用户、email 同步、firestore.MergeAll、数据流转 |
| 14:28 | 任务管理权限控制 | 资源所有权校验、环境变量开关模式、BizCode 精细化 |
| 15:59 | 请求入参校验 | go-playground/validator、Echo Validator 接口、指针接收者 error |
| 16:11 | 优化 zap logger 加载 | 按环境切换日志配置、zapcore.EncoderConfig |

---

## 1. 简化 Firebase 初始化：Auth + Firestore 合一

### 改造前

`main.go` 中分三步：先 LoadConfig → 再 InitFireStore → 再 InitFirebase，每个步骤各自处理错误，`main.go` 臃肿。

### 改造后

```go
// bootstrap/firebase.go
var AuthClient *auth.Client
var FirestoreClient *firestore.Client

func InitFirebase(ctx context.Context) {
    app, err := firebase.NewApp(context.Background(), nil, opt)
    // ...
    AuthClient, err = app.Auth(context.Background())
    // ...
    FirestoreClient, err = app.Firestore(ctx)
    // ...
}
```

```go
// main.go 简化为
bootstrap.LoadConfig()
bootstrap.InitFirebase(ctx)

taskSvc := service.NewTask(bootstrap.FirestoreClient)  // 直接用全局变量
userSvc := service.NewUser(bootstrap.FirestoreClient)
```

### 关键知识点

| 概念 | 说明 |
|------|------|
| Firebase Admin SDK | 一个 `firebase.NewApp` 实例可同时获取 Auth、Firestore、Storage 等服务 |
| 全局变量集中管理 | `AuthClient` 和 `FirestoreClient` 都在 `bootstrap` 包中，main.go 不再持有局部变量 |
| `LoadConfig()` 简化 | 从返回 `(*AppConfig, error)` 改为无返回值，因为当前只做 `godotenv.Load` |

---

## 2. 用户管理：首次登录自动创建 + Email 同步

### 核心改动：`GetUserDetailById` 的新逻辑

```
查询 Firestore 用户文档
  ├── 文档存在 → 读出 DTO → 检查 email 是否需要同步（Firebase Auth 中已修改）
  │                               └── 是 → firestore.MergeAll 仅更新 email 字段
  └── 文档不存在 → 从 UserSession 创建用户 → Set 写入 → 返回新用户
```

### 首次登录自动创建用户

```go
func (s *User) syncUserFromSession(ctx context.Context, docRef *firestore.DocumentRef) (*model.User, error) {
    session, ok := domain.FromUserSession(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "user session not found")
    }
    userDTO := dto.NewUserFromSession(session)  // 从 session 构造 DTO
    docRef.Set(ctx, userDTO)
    return userDTO.ToEntity(), nil
}
```

`dto.NewUserFromSession(session)` 从 Firebase Token 的 UID + Email 构造初始用户记录。

### Email 同步机制

```go
if session, ok := domain.FromUserSession(ctx); ok && session.Email != "" && session.Email != userDTO.Email {
    docRef.Set(ctx, map[string]any{"Email": session.Email}, firestore.MergeAll)
    userDTO.Email = session.Email
}
```

如果用户在 Firebase Auth 中修改了 email，再次查询时会自动同步到 Firestore。

### 关键知识点

| 概念 | 说明 |
|------|------|
| `firestore.MergeAll` | 只更新传入的字段，不覆盖整个文档（对比 `Set` 不带 MergeAll 是全覆盖） |
| `grpc/codes.NotFound` | Firestore 文档不存在时 gRPC 返回 NotFound，用于判断首次访问 |
| `status.Code(err)` | 从 gRPC error 中提取状态码 |

### UseCase 层从 Session 注入身份信息

```go
// usecase/user.go - CompleteUserInfo
session, ok := domain.FromUserSession(ctx)
userModel := usecaseio.ToModelUser(&input)
userModel.ID = session.UID      // ID 来自登录态，不由前端传入
userModel.Email = session.Email // Email 同理
```

`input` 中移除了 `Email` 字段 —— 身份信息不信任前端传值，从 Context 中提取。

### `dto.ToMap` 优化：零值不写入

```go
func ToMap(u *model.User) map[string]any {
    data := map[string]any{}
    if u.Username != "" {
        data["username"] = u.Username  // 只有非零值才写入
    }
    // Age 同理：if u.Age != 0 { ... }
    // Address/Profile 同理：嵌套 map 按需构建
    return data
}
```

配合 `firestore.MergeAll` 使用时，零值字段不会被错误覆盖。

---

## 3. 任务管理：权限控制（越权防护）

### 问题

任何人都可以通过 taskId 查询/修改/删除任意任务，没有归属校验。

### 解决方案：`getOwnedTask` 提取所有权限校验

```go
func (u *Task) getOwnedTask(ctx context.Context, taskId string) (*model.Task, error) {
    session, ok := domain.FromUserSession(ctx)
    if !ok {
        return nil, constants.CredentialsAbsence
    }
    task, err := u.taskSvc.GetTaskDetail(ctx, taskId)
    if err != nil {
        return nil, err
    }
    if task.Creator.ID != session.UID {
        return nil, constants.PermissionDenied  // 3005
    }
    return task, nil
}
```

`GetTaskDetail`、`ModifyTask`、`DeleteTask` 都先调用 `getOwnedTask` 做越权检查。

### 创建任务时绑定创建者

```go
session, ok := domain.FromUserSession(ctx)
task := model.Task{
    Title:   input.Title,
    Creator: model.User{ID: session.UID},  // 创建者 = 当前登录用户
}
```

### 新增 BizCode：`PermissionDenied = 3005`

```go
PermissionDenied BizCode = 3005  // "无权访问该资源" → HTTP 403
```

### BizCode → HTTP 映射从 switch 改为 map

```go
// 改造前：范围判断 switch { case c >= 3000 && c < 4000: ... }
// 改造后：精确映射
var mapBizCodeHTTPStatus = map[BizCode]int{
    Success:            http.StatusOK,
    PermissionDenied:   http.StatusForbidden,   // 403 精确控制
    CredentialsAbsence: http.StatusUnauthorized, // 401
    // ...
}
```

3000-3999 段内部也需要区分 401（未登录）和 403（已登录但无权），范围判断不再满足。

### 邮箱验证开关

```go
var requireEmailVerify = false  // 默认不校验

func (fa *FirebaseAuthenticator) Authenticate(...) {
    require, err := strconv.ParseBool(os.Getenv("REQUIRE_EMAIL_VERIFY"))
    if err != nil {
        require = true  // 读不到配置时默认开启（安全默认值）
    }
    if (!emailVerified) && require {
        return nil, constants.EmailUnverified
    }
}
```

`REQUIRE_EMAIL_VERIFY=true` 才校验邮箱，本地开发设为 `false` 方便调试。

### 关键知识点

| 概念 | 说明 |
|------|------|
| 资源所有权校验 | 用 `creator_id` 匹配 `session.UID`，防止越权 |
| 安全默认值 | 环境变量读取失败时默认启用最严格的安全策略 |
| `strconv.ParseBool` | 字符串 "true"/"false" → bool |
| 精细化 HTTP 状态码 | 401 vs 403 语义不同，不能在范围判断中混为一谈 |

---

## 4. 请求入参校验：go-playground/validator

### 为什么引入

之前请求 DTO 没有校验，空 title、超长字符串、负数 age 都能直接打到 usecase 层。

### 架构

```
Request DTO struct tag  →  go-playground/validator  →  CustomValidator
       ↓                                                     ↓
  validate:"required,min=1,max=200"               ValidationError{BizCode, Message}
                                                          ↓
                                              errors.AsType[*ValidationError]
                                                    在 error_handler / Fail 中匹配
```

### 实现 Echo 的 Validator 接口

```go
// internal/validator/custom.go
type CustomValidator struct {
    v *govalidator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
    if err := cv.v.Struct(i); err != nil {
        var errs govalidator.ValidationErrors
        if errors.As(err, &errs) && len(errs) > 0 {
            return &ValidationError{
                BizCode: constants.InvalidInputParam,
                Message: translate(errs[0]),  // "Title 不能为空"
            }
        }
        // ...
    }
    return nil
}
```

注册到 Echo：

```go
e.Validator = customvalidator.New()
```

### Handler 中调用

```go
if err := c.Validate(req); err != nil {
    return reponse.Fail(c, err)
}
```

### 错误信息翻译

```go
func translate(fe govalidator.FieldError) string {
    switch fe.Tag() {
    case "required": return fmt.Sprintf("%s 不能为空", field)
    case "min":      return fmt.Sprintf("%s 不能小于 %s", field, fe.Param())
    case "max":      return fmt.Sprintf("%s 不能大于 %s", field, fe.Param())
    }
}
```

### 请求 DTO 示例

```go
type CreateTask struct {
    Title       string `json:"title" validate:"required,min=1,max=200"`
    Description string `json:"description" validate:"max=200000000"`
}

type CompleteUserInfoInput struct {
    Username string `json:"username" validate:"required,min=1,max=50"`
    Age      int    `json:"age" validate:"min=0,max=200"`
}
```

### 关键知识点

| 概念 | 说明 |
|------|------|
| Echo `Validator` 接口 | 实现 `Validate(interface{}) error`，通过 `e.Validator = xxx` 注册 |
| `c.Validate(req)` | Echo 自动在绑定的 struct 上调用 Validator |
| 指针接收者的 `*ValidationError` | `errors.AsType[*ValidationError](err)` 必须传指针类型 |
| struct tag 声明式校验 | `validate:"required,min=1,max=200"` 规则写在 tag 中 |
| `errors.As` vs `errors.AsType` | `As` 是标准库（Go 1.13），`AsType` 是 Go 1.25 泛型版，此处用了标准库版 |

### 注意事项（代码中的注释）

> `ValidationError` 的方法是**指针接收者**，所以 `errors.AsType` 传入 `*ValidationError` 而非 `ValidationError`。这与 BizCode（值接收者 → 值类型）相反。

---

## 5. Zap Logger：按环境自适应配置

### 改造前

```go
logger, _ := zap.NewProduction()  // 所有环境都用 Production 配置
```

### 改造后

```go
func InitLogger() (*zap.Logger, error) {
    env := os.Getenv("APP_ENV")
    switch strings.ToLower(env) {
    case "local", "dev", "development", "test":
        // 开发环境：console 输出、Debug 级别、彩色
        config := zap.NewDevelopmentConfig()
        config.Encoding = "console"
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

    case "staging", "stage":
        // 预发布：JSON 输出、Debug 级别（便于排查）

    case "production", "prod":
        // 生产：JSON 输出、Info 级别（减少日志量）
    }
}
```

### 不同环境的差异

| 环境 | Encoding | Level | 特点 |
|------|----------|-------|------|
| local/dev/test | console | Debug | 人类可读、彩色、所有日志 |
| staging | json | Debug | 机器可读、保留 Debug 方便排查 |
| production | json | Info | 机器可读、不输出 Debug 减少 I/O |

### 为什么 console 用 `zap.NewDevelopmentConfig()` 而不是 `NewProductionConfig()`

`NewDevelopmentConfig` 默认开启 stacktrace、caller 信息，本地调试更友好。再覆盖 `encoding="console"` 转为人类可读格式。

### 关键知识点

| 概念 | 说明 |
|------|------|
| `zap.NewDevelopmentConfig()` | 开发配置：Debug 级别、stacktrace、caller 行号 |
| `zap.NewProductionConfig()` | 生产配置：Info 级别、JSON、性能优化 |
| `zapcore.CapitalColorLevelEncoder` | 彩色日志级别（DEBUG=蓝 INFO=绿 ERROR=红） |
| `zapcore.ISO8601TimeEncoder` | 时间格式 `2025-05-27T16:11:57+0800` |
| `zap.NewAtomicLevelAt(level)` | 动态日志级别（运行时可通过 HTTP 端点动态调整） |

---

## 今日 Go 语言知识点速查表

| 知识点 | 出现位置 |
|--------|----------|
| Firebase Admin SDK 多服务初始化 | `firebase.NewApp` → `app.Auth` + `app.Firestore` |
| `firestore.MergeAll` | 部分字段更新（用户 email 同步） |
| `grpc/codes.NotFound` + `status.Code(err)` | Firestore 文档不存在的判断 |
| `strconv.ParseBool` | 环境变量开关（邮箱验证开关） |
| 资源所有权校验模式 | `getOwnedTask` 统一入口 |
| 安全默认值 (fail-secure) | 读不到配置时默认启用邮箱验证 |
| BizCode 精细化 → map 映射 | `PermissionDenied` → 403 |
| go-playground/validator | struct tag 声明式校验 |
| Echo `Validator` 接口 | `Validate(interface{}) error` |
| `errors.As` 指针接收者 | `&govalidator.ValidationErrors` 和 `*ValidationError` |
| `zapcore.EncoderConfig` | 自定义日志时间格式、颜色、encoding |
| 按环境切换配置 | `APP_ENV` 环境变量驱动的 switch |

---

## 下一步建议

1. **validator 国际化** — 当前 `translate` 硬编码中文，可改为根据 `Accept-Language` header 返回多语言错误
2. **validator 单元测试** — 构造各种非法 struct 测试 `CustomValidator.Validate`
3. **权限校验测试** — 用不同 UID 的 UserSession 测试 `getOwnedTask` 是否能正确拒绝
4. **logger 热重载** — 利用 `zap.AtomicLevel` 通过 HTTP 端点动态切换日志级别，无需重启
5. **Firestore batch** — 如果创建用户需要同时写多个文档，考虑用 `firestore.BulkWriter` 或 `Batch` 保证原子性
