package apperror

import (
	"errors"
	"go-echo-demo/internal/domain"
	"net/http"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func MapError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return &AppError{
			Status:  http.StatusNotFound,
			Code:    "not found",
			Message: "not found",
			Err:     err,
		}
	case errors.Is(err, domain.ErrForbidden):
		return &AppError{
			Status:  http.StatusForbidden,
			Code:    "forbidden",
			Message: "forbidden",
			Err:     err,
		}
	case errors.Is(err, domain.ErrInvalidInput):
		return &AppError{
			Status:  http.StatusBadRequest,
			Code:    "invalid input",
			Message: "invalid input",
			Err:     err,
		}
	default:
		return &AppError{
			Status:  http.StatusInternalServerError,
			Code:    "internal server error",
			Message: "internal server error",
			Err:     err,
		}
	}

}
