package constants

import "net/http"

type BizCode int

// 定义业务状态码 大概按照以下分段规则来进行
// ┌───────────┬─────────────────────┐
// │   范围     │        含义          │
// ├───────────┼─────────────────────┤
// │ 0         │ 成功                 │
// ├───────────┼─────────────────────┤
// │ 1000-1999 │ 参数校验 / 输入错误    │
// ├───────────┼─────────────────────┤
// │ 2000-2999 │ 资源不存在            │
// ├───────────┼─────────────────────┤
// │ 3000-3999 │ 权限 / 鉴权          │
// ├───────────┼─────────────────────┤
// │ 4000-4999 │ 业务规则冲突          │
// ├───────────┼─────────────────────┤
// │ 9000-9999 │ 系统内部错误          │
// └───────────┴─────────────────────┘
// 业务状态码表

const (
	Success            BizCode = 0
	InvalidInputParam  BizCode = 1001
	RequireAbsence     BizCode = 1002
	UnknownScene       BizCode = 1003
	UserNotFound       BizCode = 2001
	TaskNotFound       BizCode = 2002
	EmailUnverified    BizCode = 3001
	CredentialsAbsence BizCode = 3002
	CredentialsExpired BizCode = 3003
	TokenInvalid       BizCode = 3004
	PermissionDenied   BizCode = 3005
	InvalidCursor      BizCode = 4001
	DocMapError        BizCode = 4002
	TaskNotArchivable  BizCode = 4003 // 任务状态不满足归档条件（必须为已完成）
	RequireTransaction BizCode = 4004 // 需要事务 但是没从context中找到对应标志
	InternalError      BizCode = 9001
)

var mapBizCodeMsg = map[BizCode]string{
	InvalidInputParam:  "参数不合法！",
	UserNotFound:       "用户不存在！",
	RequireAbsence:     "缺失必要的参数",
	TaskNotFound:       "任务不存在！",
	EmailUnverified:    "邮箱未验证！",
	InternalError:      "系统内部错误！",
	CredentialsAbsence: "凭证缺失或格式不符",
	CredentialsExpired: "凭证过期",
	TokenInvalid:       "无效Token",
	PermissionDenied:   "无权访问该资源",
	InvalidCursor:      "非法游标",
	UnknownScene:       "无法满足的查询请求",
	DocMapError:        "执行转换时出错", // 可能要改一下这里
	TaskNotArchivable:  "只有已完成的任务才能被归档",
	RequireTransaction: "需要事务控制",
}

// 1. 实现 Go 的 error 接口
func (c BizCode) Error() string {
	if msg, ok := mapBizCodeMsg[c]; ok {
		return msg
	}
	return "未知错误"
}

// 建立BizCode到HTTP状态码的映射
var mapBizCodeHTTPStatus = map[BizCode]int{
	Success:            http.StatusOK,
	InvalidInputParam:  http.StatusBadRequest,
	RequireAbsence:     http.StatusBadRequest,
	UnknownScene:       http.StatusBadRequest,
	DocMapError:        http.StatusInternalServerError,
	UserNotFound:       http.StatusNotFound,
	TaskNotFound:       http.StatusNotFound,
	EmailUnverified:    http.StatusUnauthorized,
	CredentialsAbsence: http.StatusUnauthorized,
	CredentialsExpired: http.StatusUnauthorized,
	TokenInvalid:       http.StatusUnauthorized,
	PermissionDenied:   http.StatusForbidden,
	InternalError:      http.StatusInternalServerError,
	InvalidCursor:      http.StatusBadRequest,
	TaskNotArchivable:  http.StatusUnprocessableEntity,
	RequireTransaction: http.StatusInternalServerError,
}

// HTTPStatus 返回业务码对应的 HTTP 状态码
func (c BizCode) HTTPStatus() int {
	if status, ok := mapBizCodeHTTPStatus[c]; ok {
		return status
	}
	return http.StatusInternalServerError
}
