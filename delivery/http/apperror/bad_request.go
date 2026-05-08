package apperror

import "net/http"

func BadRequest(code, message string, err error) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Code:    code,
		Message: message,
		Err:     err,
	}
}
