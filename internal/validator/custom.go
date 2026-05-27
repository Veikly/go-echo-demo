package validator

import (
	"errors"
	"fmt"

	"go-echo-demo/internal/constants"

	govalidator "github.com/go-playground/validator/v10"
)

// ValidationError 携带详情信息的参数校验错误
type ValidationError struct {
	BizCode constants.BizCode
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// HTTPStatus 代理到内部 BizCode 的 HTTP 状态码
func (e *ValidationError) HTTPStatus() int {
	return e.BizCode.HTTPStatus()
}

type CustomValidator struct {
	v *govalidator.Validate
}

func New() *CustomValidator {
	return &CustomValidator{v: govalidator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.v.Struct(i); err != nil {
		var errs govalidator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			return &ValidationError{
				BizCode: constants.InvalidInputParam,
				Message: translate(errs[0]),
			}
		}
		return &ValidationError{
			BizCode: constants.InvalidInputParam,
			Message: "参数校验失败",
		}
	}
	return nil
}

func translate(fe govalidator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s 不能为空", field)
	case "min":
		return fmt.Sprintf("%s 不能小于 %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s 不能大于 %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s 校验失败: %s", field, fe.Tag())
	}
}
