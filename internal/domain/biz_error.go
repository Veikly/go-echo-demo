package domain

import (
	"errors"
	"go-echo-demo/internal/constants"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
)

type BizError struct {
	HTTPStatus int
	BizCode    constants.BizCode
	Message    string
}
