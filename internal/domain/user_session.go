package domain

import "context"

type UserSession struct {
	UID           string
	Email         string
	EmailVerified bool
}

// 统一认证接口
type Authenticator interface {
	Authenticate(ctx context.Context, tokenString string) (*UserSession, error)
}
