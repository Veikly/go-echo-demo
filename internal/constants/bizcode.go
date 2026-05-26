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
	UserNotFound       BizCode = 2001
	TaskNotFound       BizCode = 2002
	EmailUnverified    BizCode = 3001
	CredentialsAbsence BizCode = 3002
	CredentialsExpired BizCode = 3003
	TokenInvalid       BizCode = 3004
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
}

// 1. 实现 Go 的 error 接口
func (c BizCode) Error() string {
	if msg, ok := mapBizCodeMsg[c]; ok {
		return msg
	}
	return "未知错误"
}

// 2. 建立业务码到 HTTP 状态码的映射
func (c BizCode) HTTPStatus() int {
	switch {
	case c == Success:
		return http.StatusOK
	case c >= 1000 && c < 2000:
		return http.StatusBadRequest // 400
	case c >= 2000 && c < 3000:
		return http.StatusNotFound // 404
	case c >= 3000 && c < 4000:
		return http.StatusUnauthorized // 401 或 403
	case c >= 4000 && c < 5000:
		return http.StatusConflict // 409
	default:
		return http.StatusInternalServerError // 500
	}
}
