# Go + Echo + Clean Architecture 学习总结 (2025-05-26)

## 今日提交总览

| 时间 | 提交 | 知识点 |
|------|------|--------|
| 13:54 | 认证中间件重构 | 接口抽象、依赖倒置、Context 传值、中间件工厂 |
| 15:14 | Handler 层架构泄漏修复 | 响应 DTO 隔离、层级边界、DTO→Entity 转换 |
| 18:09 | 异常处理体系与全局响应 | BizCode 错误码、errors.AsType、统一 ApiResponse |
| 18:09 | 补充提交 | 响应格式修正 |

---

## 1. 依赖倒置：认证中间件依赖于接口而非具体实现

### 改造前（架构泄漏）

```go
// ❌ 中间件直接依赖 bootstrap 包的全局变量
func FirebaseAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        token, err := bootstrap.AuthClient.VerifyIDToken(...)
        // ...
        c.Set("firebase_token", token)   // Firebase 概念泄漏到 handler 层
    }
}
```

问题：
- 中间件直接依赖 `bootstrap.AuthClient` 全局变量，无法替换认证方式
- `c.Set("firebase_token", token)` 把 Firebase 专有类型暴露给了 handler
- 测试时无法 mock 认证逻辑

### 改造后（面向接口）

**Step 1 — 在 domain 层定义接口 + 数据模型**

```go
// internal/domain/user_session.go
type UserSession struct {
    UID           string
    Email         string
    EmailVerified bool
}

type Authenticator interface {
    Authenticate(ctx context.Context, tokenString string) (*UserSession, error)
}
```

**Step 2 — Firebase 实现该接口**

```go
// internal/infra/authenticator/firebase_auth.go
type FirebaseAuthenticator struct {
    authClient *auth.Client
}

func (fa *FirebaseAuthenticator) Authenticate(ctx context.Context, tokenString string) (*domain.UserSession, error) {
    token, err := fa.authClient.VerifyIDToken(ctx, tokenString)
    // ...
    return &domain.UserSession{UID: uid, Email: email}, nil
}
```

**Step 3 — 中间件通过构造函数接收接口**

```go
// delivery/http/appmiddleware/auth_middleware.go
func NewAuthMiddleware(authenticator domain.Authenticator) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            userSession, err := authenticator.Authenticate(c.Request().Context(), idToken)
            // 注入 UserSession 到 Context（而非 Firebase Token）
            reqCtx := domain.WithUserSession(c.Request().Context(), *userSession)
            c.SetRequest(c.Request().WithContext(reqCtx))
            return next(c)
        }
    }
}
```

**Step 4 — main.go 中组装依赖**

```go
firebaseAuthenticator := authenticator.NewFirebaseAuthenticator(bootstrap.AuthClient)
authMiddleware := appmiddleware.NewAuthMiddleware(firebaseAuthenticator)
e.Use(authMiddleware)
```

### 关键知识点

| 概念 | 说明 |
|------|------|
| 依赖倒置原则 (DIP) | 高层模块（中间件）不依赖低层模块（Firebase），两者都依赖抽象（Authenticator 接口） |
| 接口隔离 (ISP) | Authenticator 只有一个方法，最小化接口 |
| 工厂函数 | `NewAuthMiddleware(authenticator)` 通过参数注入依赖，而非内部创建 |
| Context 传值 | `domain.WithUserSession(ctx, session)` + `domain.FromUserSession(ctx)` 替代裸 `c.Set/Get` |
| 类型安全的 Context Key | `type userContextKey string` 防止 key 冲突 |

---

## 2. 架构泄漏修复：响应 DTO 与实体隔离

### 问题：Handler 直接返回 domain/usecaseio 对象

```go
// ❌ 之前：usecase/usecaseio 的 struct 直接作为 JSON 返回
return c.JSON(http.StatusOK, detail)  // detail 是 usecaseio.TaskDetailOutput

// ❌ 之前：Firestore DTO 直接返回给 usercase
var u model.User
return &u, nil  // 应该是 DTO → Entity 转换
```

### 修复

**Handler 层添加响应 DTO 转换**

```go
// internal/handler/task.go
rsp := response.TaskDetail{
    ID:          detail.ID,
    Title:       detail.Title,
    Description: detail.Description,
    // ...
}
return reponse.Success(c, rsp)
```

**新增 `internal/response/` 包**

```
internal/response/
├── task.go   → SaveTask, TaskDetail
└── user.go   → Address, Profile, UserDetail, CompleteUserInfo
```

**修复 Firestore 层 DTO → Entity 转换**

```go
// ❌ 之前
var u model.User
data.DataTo(&u)
return &u, nil

// ✅ 修复
var u dto.User
data.DataTo(&u)
u.ID = docRef.ID
return u.ToEntity(), nil
```

### 层级数据流

```
Request      → internal/request/    (请求 DTO)
Handler      → internal/response/   (响应 DTO)  ← 今日新增
UseCase      → internal/usecase/usecaseio/  (业务 I/O)
Model        → internal/model/      (领域模型)
Infra DTO    → internal/infra/.../dto/  (持久化 DTO)
```

每一层使用自己的数据结构，不跨层泄漏类型。

---

## 3. BizCode 自定义业务状态码体系

### 设计思路

用 `type BizCode int` 统一管理业务状态码，把"发生了什么错误"和"该返回什么 HTTP 状态码"解耦。

### 核心代码

```go
// internal/constants/bizcode.go
type BizCode int

// 分段定义
// 0:         成功
// 1000-1999: 参数校验 / 输入错误
// 2000-2999: 资源不存在
// 3000-3999: 权限 / 鉴权
// 4000-4999: 业务规则冲突
// 9000-9999: 系统内部错误

const (
    Success            BizCode = 0
    InvalidInputParam  BizCode = 1001
    RequireAbsence     BizCode = 1002
    UserNotFound       BizCode = 2001
    TaskNotFound       BizCode = 2002
    EmailUnverified    BizCode = 3001
    CredentialsAbsence BizCode = 3002
    CredentialsExpired BizCode = 3003
    TokenInvalid       BizCode = 3004
    InternalError      BizCode = 9001
)

// 实现 error 接口 → BizCode 本身可以当 error 用
func (c BizCode) Error() string {
    return mapBizCodeMsg[c]
}

// 映射到 HTTP 状态码
func (c BizCode) HTTPStatus() int {
    switch {
    case c >= 1000 && c < 2000: return http.StatusBadRequest    // 400
    case c >= 2000 && c < 3000: return http.StatusNotFound       // 404
    case c >= 3000 && c < 4000: return http.StatusUnauthorized   // 401
    case c >= 4000 && c < 5000: return http.StatusConflict      // 409
    default:                    return http.StatusInternalServerError // 500
    }
}
```

### Go 1.25 `errors.AsType` 泛型用法

```go
// errors.AsType[T] 是 Go 1.25 新增的泛型版 errors.As
// 等效于之前的 errors.As，但不需要传入指针参数

if c, ok := errors.AsType[constants.BizCode](err); ok {
    httpStatus = c.HTTPStatus()
    bizCode = c
    msg = c.Error()
} else {
    // 非 BizCode 的错误（如 DB 原生错误）→ 500 + 通用提示
    msg = "系统内部错误！"
}
```

### 注意事项（代码中的注释）

> 如果自定义错误类型的方法是**值接收者**，`AsType` 传值和传指针都支持。如果是**指针接收者**，必须传指针，否则类型判断无效。

BizCode 的 `Error()` 和 `HTTPStatus()` 都是值接收者，所以 `errors.AsType[BizCode](err)` 没问题。

---

## 4. 统一响应格式 ApiResponse

### 格式定义

```go
// delivery/http/reponse/response.go
type ApiResponse struct {
    Code    constants.BizCode `json:"code"`
    Message string            `json:"message"`
    Data    any               `json:"data"`
}
```

之前的状态：
```
❌ 成功时：直接返回业务 JSON
❌ 失败时：返回 ErrorResponse{code, message}
```
现在统一为：
```
✅ 全部走 ApiResponse{code, message, data}
   成功：{code: 0, message: "", data: {...}}
   失败：{code: 2001, message: "用户不存在！", data: null}
```

### 两个辅助函数

```go
// handler 成功时调用
func Success(c echo.Context, data any) error {
    return c.JSON(http.StatusOK, ApiResponse{
        Code:    constants.Success,
        Data:    data,
    })
}

// handler 失败时调用
func Fail(c echo.Context, err error) error {
    // 用 errors.AsType 判断是否为 BizCode
    // → 是：取对应的 HTTPStatus 和错误信息
    // → 否：返回 500 + "系统内部错误！"
    return c.JSON(httpStatus, ApiResponse{...})
}
```

### 全局错误处理器中也用同一套格式

```go
// appmiddleware/error_handler.go
func CustomHTTPErrorHandler(err error, c echo.Context) {
    // 同样用 errors.AsType[BizCode] 判断
    // 同样返回 ApiResponse 格式
}
```

---

## 5. UseCase 层全面改用 BizCode

```go
// ❌ 之前
return usecaseio.CreateTaskOutput{}, domain.ErrInvalidInput
return usecaseio.UserDetailOutput{}, errors.New("userId can't be null")

// ✅ 现在
return usecaseio.CreateTaskOutput{}, constants.RequireAbsence
return usecaseio.UserDetailOutput{}, constants.InvalidInputParam
```

UseCase 和 Infra Service 层不再返回原始 `errors.New` 或 `domain.ErrXxx`，统一返回 `constants.BizCode`。

---

## 今日 Go 语言知识点速查表

| 知识点 | 出现位置 |
|--------|----------|
| 依赖倒置原则 (DIP) | Authenticator 接口 + 中间件工厂 |
| 接口定义与实现 | `domain.Authenticator` → `FirebaseAuthenticator` |
| 构造函数依赖注入 | `NewAuthMiddleware(authenticator)` |
| Context 传值（类型安全的 key） | `WithUserSession` / `FromUserSession` |
| `errors.AsType[T]`（Go 1.25） | BizCode 类型判断 |
| `type BizCode int` 带方法 | 业务状态码实现 error 接口 |
| 分段常量（1000/2000/3000/9000） | 状态码范围设计 |
| 统一响应 DTO 层 | `internal/response/` |
| 值接收者 vs 指针接收者 | `errors.AsType` 的类型判断细节 |
| `c.SetRequest(c.Request().WithContext(...))` | 把值注入 HTTP request 的 Context |
| Echo MiddlewareFunc 签名 | `func(echo.HandlerFunc) echo.HandlerFunc` |

---

## 下一步建议

1. **单元测试** — 给 `reponse.Fail()` 写 table-driven test，验证不同 BizCode → HTTPStatus 的映射
2. **自定义 error 类型进阶** — 尝试用指针接收者 + `errors.AsType`，对比值接收者的行为差异
3. **Middleware 单元测试** — mock Authenticator 接口来测试中间件的各种分支（空 header、格式错误、token 无效）
4. **国际化** — BizCode 的错误信息可以从常量改为支持 i18n 的消息函数
