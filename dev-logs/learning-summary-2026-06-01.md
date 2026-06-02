# Go + Echo + Clean Architecture 学习总结 (2026-06-01)

## 今日提交总览

| 时间 | 提交 | 知识点 |
|------|------|--------|
| 14:12 | 分页查询重构：provider 模式 + 错误码统一 | 依赖装配下沉、链式挂载、BizCode 替换原生 error、删除双向游标 |
| 15:32 | 完善查询记录总数功能 | `WithTotalCount` 参数透传、`IncludeTotalCount` 字段补全 |
| 16:42 | 分页查询全模块单元测试 | fake repo、table-driven、纯函数测试、`errors.Is` 断言 |

---

## 1. Provider 模式：把装配逻辑从 main.go 中剥离

### 问题：main.go 越来越臃肿

随着分页查询功能的加入，`main.go` 里出现了大量装配代码：创建 repo、创建 registry、注册 scene、构造 usecase、注入 injectRules……这些都是"怎么组装"的细节，不是"启动流程"本身。

### 解决方案：`bootstrap/provider` 包

新增 `internal/bootstrap/provider/task.go`，把 task 相关的所有组件装配封装成一个工厂函数：

```go
// internal/bootstrap/provider/task.go
func NewTaskHandler(client *firestore.Client) *handler.TaskHandler {
    // 基础 CRUD
    taskSvc := service.NewTask(client)
    taskUseCase := usecase.NewTask(taskSvc)

    // 分页查询
    registry := dmpagination.NewRegistry()
    scene.RegisterTaskScenes(registry)

    repo := fspagination.NewFirestoreRepository[model.Task](client, "tasks", mapper.TaskMapper)

    uc := ucpagination.NewQueryUseCase(ucpagination.QueryUseCaseConfig[model.Task, response.TaskItem]{
        Repo:     repo,
        Registry: registry,
        ToDTO:    func(t model.Task) response.TaskItem { ... },
        InjectRules: func(ctx context.Context, q dmpagination.PageQuery) (dmpagination.PageQuery, error) {
            // 注入 creator_id 过滤，只允许访问自己创建的任务
            session, ok := domain.FromUserSession(ctx)
            if !ok {
                return q, constants.CredentialsAbsence
            }
            q.Filters = append(q.Filters, dmpagination.FilterCriteria{
                Field: "creator_id", Op: dmpagination.FilterOpEq, Value: session.UID,
            })
            return q, nil
        },
    })

    listHandler := handler.NewTaskListHandler(uc, registry)
    return handler.NewTask(taskUseCase).WithListHandler(listHandler)
}
```

`main.go` 简化为：

```go
// 改造前：main.go 里有 50+ 行装配代码
// 改造后：
server := bootstrap.Server{
    Echo:        e,
    TaskHandler: provider.NewTaskHandler(bootstrap.FirestoreClient),
    UserHandler: provider.NewUserHandler(bootstrap.FirestoreClient),
}
```

### 关键知识点

| 概念 | 说明 |
|------|------|
| Provider / Factory 模式 | 把"如何创建一组相关对象"封装成一个函数，调用方只关心结果 |
| 单一职责 | `main.go` 只负责启动流程，装配细节下沉到 provider |
| 依赖注入入口 | `*firestore.Client` 是唯一的外部依赖，从 bootstrap 全局变量传入 |

---

## 2. 链式挂载：分页 Handler 不再是独立 Handler

### 改造前

`Server` 结构体里有一个独立的 `TaskPageHandler` 字段，路由绑定时单独注册：

```go
// 改造前
type Server struct {
    TaskHandler     *handler.TaskHandler
    TaskPageHandler handler.TaskPageHandlerFunc  // 独立字段
}

// 路由绑定
taskGroup.GET("", server.TaskPageHandler)
```

### 改造后

分页查询作为 `TaskHandler` 的可选能力，通过链式调用挂载：

```go
// 改造后
type Server struct {
    TaskHandler *handler.TaskHandler  // 只有一个字段
}

// provider 层装配
return handler.NewTask(taskUseCase).WithListHandler(listHandler)

// 路由绑定
taskGroup.GET("", server.TaskHandler.ListTasks)
```

`WithListHandler` 是 `TaskHandler` 上的方法，返回 `*TaskHandler` 自身，支持链式调用。

### 为什么这样更好

- `Server` 结构体不再感知"分页"这个实现细节，只知道有一个 `TaskHandler`
- 分页能力是 task 功能的一部分，放在 `TaskHandler` 内部更内聚
- 如果某个资源不需要分页，不调用 `WithListHandler` 即可，不影响其他逻辑

---

## 3. 错误码统一：用 BizCode 替换原生 error

### 改造前

分页模块的错误散落在各处，用的是原生 `fmt.Errorf` 或 `echo.NewHTTPError`：

```go
// scene/task.go 改造前
return dmpagination.PageQuery{}, fmt.Errorf("status is required for scene task.by_status_title")
return dmpagination.PageQuery{}, fmt.Errorf("invalid status value: %q", v)

// page_query.go 改造前
return echo.NewHTTPError(http.StatusBadRequest, "scene is required")
return echo.NewHTTPError(http.StatusBadRequest, "unknown scene: "+base.Scene)
```

### 改造后

全部替换为 `constants.BizCode`：

```go
// scene/task.go 改造后
return dmpagination.PageQuery{}, constants.RequireAbsence
return dmpagination.PageQuery{}, constants.InvalidInputParam

// page_query.go 改造后
return constants.RequireAbsence
return constants.UnknownScene
return constants.InvalidCursor
```

### 为什么这样更好

- 错误信息和 HTTP 状态码的映射集中在 `bizcode.go` 一处维护，不散落在各层
- `BizCode` 实现了 `error` 接口，可以直接 `return`，也可以被 `errors.Is` 匹配
- delivery 层（`page_query.go`）不再直接依赖 `echo`，降低了 HTTP 框架的侵入性
- 新增了 `UnknownScene = 1003` 和 `InvalidCursor = 4001` 两个业务码

---

## 4. 删除双向游标：简化分页模型

### 删除的内容

```go
// 删除前：domain/pagination/page.go
type CursorDir string

const (
    CursorForward  CursorDir = "forward"
    CursorBackward CursorDir = "backward"
    CursorRefresh  CursorDir = "refresh"
)

type PageQuery struct {
    Direction CursorDir  // 删除
    // ...
}

type PageResult[T any] struct {
    PrevCursor string  // 删除
    // ...
}
```

### 原因

当前 Firestore 的游标实现只支持向前翻页（`StartAfter`），双向游标是过度设计。删除 `Direction`、`PrevCursor` 后，`ApplyPaging` 的签名也随之简化：

```go
// 改造前
func ApplyPaging(base PageQuery, rawCursor string, dir CursorDir, limit int) (PageQuery, error)

// 改造后
func ApplyPaging(base PageQuery, rawCursor string, limit int) (PageQuery, error)
```

YAGNI 原则（You Aren't Gonna Need It）：不要为"将来可能需要"的功能提前设计，等真正需要时再加。

---

## 5. 完善查询记录总数功能

### 问题

`WithTotalCount` 参数在 HTTP 层（`BasePageQuery`）和 UseCase 层（`ExecuteInput`）都没有定义，导致前端传了 `with_total_count=true` 也不会生效。

### 修复

三处同步补全：

```go
// 1. delivery/http/pagination/page_query.go
type BasePageQuery struct {
    Scene          string `query:"scene"`
    Cursor         string `query:"cursor"`
    Limit          int    `query:"limit"`
    WithTotalCount bool   `query:"with_total_count"`  // 新增
}

// 2. internal/usecase/pagination/page.go
type ExecuteInput struct {
    Scene          pagination.SceneID
    Params         pagination.SceneParams
    Cursor         string
    Limit          int
    WithTotalCount bool  // 新增
}

// 3. Execute 方法中透传
q.IncludeTotalCount = input.WithTotalCount  // 新增
```

### 数据流

```
HTTP 请求 ?with_total_count=true
    → BasePageQuery.WithTotalCount = true
    → ExecuteInput.WithTotalCount = true
    → PageQuery.IncludeTotalCount = true
    → FirestoreRepository 执行 AggregationQuery
    → PageResult.TotalCount = &count
    → 响应 JSON 中包含 total_count 字段
```

---

## 6. 分页查询全模块单元测试

今天完成了分页查询链路上所有可测模块的单元测试，共 6 个测试文件，50+ 个用例，全部通过。

### 测试文件清单

```
delivery/http/pagination/page_query_test.go       → ValidateBaseParams（10 个用例）
internal/domain/pagination/cursor_test.go         → EncodeCursor/DecodeCursor（10 个用例）
internal/domain/pagination/page_test.go           → ApplyPaging（7 个用例）
internal/domain/pagination/scene_registry_test.go → Registry（5 个用例）
internal/infra/firestore/scene/task_test.go       → scene builder（18 个用例）
internal/usecase/pagination/page_test.go          → QueryUseCase.Execute（10 个用例）
```

### 核心测试模式：fake repo

`FirestoreRepository` 强依赖 Firestore SDK，无法在单测中使用。用 fake 实现替代：

```go
type fakeTaskRepo struct {
    result    pagination.PageResult[model.Task]
    err       error
    lastQuery pagination.PageQuery  // 记录收到的 query，供断言使用
}

func (r *fakeTaskRepo) Query(_ context.Context, q pagination.PageQuery) (pagination.PageResult[model.Task], error) {
    r.lastQuery = q
    return r.result, r.err
}
```

`lastQuery` 字段是关键设计：不仅能验证返回值，还能断言"传入 repo 的 query 是否正确"，比如 cursor 是否被解码、injectRules 是否生效、limit 是否被正确传递。

### 错误断言：`errors.Is` vs `==`

`BizCode` 是值类型，实现了 `error` 接口，`errors.Is` 内部会做 `==` 比较：

```go
// 两种写法等价，推荐用 errors.Is（更符合 Go 惯例，未来加 wrapping 也能工作）
if err != constants.InvalidCursor { ... }
if !errors.Is(err, constants.InvalidCursor) { ... }
```

### 手动构造非法 cursor 测试版本校验

`cursorVersion` 是包内私有常量，测试中无法直接引用。通过手动构造 JSON payload 绕过：

```go
func TestDecodeCursor_VersionMismatch(t *testing.T) {
    payload := map[string]any{
        "id": "doc1", "sf": "updated_at",
        "sv": "0", "svt": "int64",
        "ver": 99,  // 与内部 cursorVersion=1 不符
    }
    b, _ := json.Marshal(payload)
    encoded := base64.URLEncoding.EncodeToString(b)

    _, err := DecodeCursor(encoded)
    if err != constants.InvalidCursor {
        t.Errorf("expected constants.InvalidCursor, got %v", err)
    }
}
```

这个技巧适用于所有"需要测试内部校验逻辑但无法直接构造非法状态"的场景。

### scene builder 测试：`toTaskStatus` 的多类型分支

`toTaskStatus` 支持 4 种输入类型，每种都要测：

```go
// enums.TaskStatus 类型（直接传枚举值）
buildTaskByStatusTitle(SceneParams{"status": enums.StatusTodo})

// int 类型
buildTaskByStatusTitle(SceneParams{"status": int(1)})

// float64 类型（JSON 解析默认产生 float64，实际请求中最常见）
buildTaskByStatusTitle(SceneParams{"status": float64(2)})

// string 类型（URL query 参数）
buildTaskByStatusTitle(SceneParams{"status": "1"})
```

float64 分支是最容易被遗漏的，但在实际 HTTP 请求中 JSON body 里的数字默认就是 float64。

---

## 今日 Go 语言知识点速查表

| 知识点 | 出现位置 |
|--------|----------|
| Provider / Factory 模式 | `bootstrap/provider/task.go` |
| 链式方法 `WithXxx()` 返回自身 | `handler.NewTask(...).WithListHandler(...)` |
| YAGNI 原则 | 删除双向游标 `CursorDir`、`PrevCursor` |
| `BizCode` 替换 `echo.NewHTTPError` | `page_query.go`、`scene/task.go` |
| fake 接口实现（测试替身） | `fakeTaskRepo` 记录 `lastQuery` |
| `errors.Is` 对值类型 error 的匹配 | cursor/scene 错误断言 |
| `base64.URLEncoding.EncodeToString` | 手动构造非法 cursor payload |
| `json.Marshal` 构造测试数据 | 版本号不匹配的 cursor 测试 |
| 参数透传的完整链路 | `WithTotalCount` 从 HTTP 到 Firestore |
| `float64` 是 JSON 数字的默认类型 | `toTaskStatus` 的 float64 分支 |

---

## 下一步建议

1. **handler 层测试** — 用 `httptest` 包测试 `PaginatedHandler`，验证 HTTP 请求到响应的完整链路
2. **`WithListHandler` 的 nil 防护** — 如果 `listHandler` 为 nil，`ListTasks` 应该返回 405 而不是 panic
3. **provider 层测试** — `NewTaskHandler` 目前没有测试，可以考虑用 Firestore Emulator 做集成测试
4. **cursor 签名防篡改** — `cursor.go` 里有 `todo 对 cursor 进行签名防篡改`，可以用 HMAC-SHA256 实现
5. **分页响应缓存** — 对于不频繁变化的列表，可以在 handler 层加 `Cache-Control` header 减少重复查询
