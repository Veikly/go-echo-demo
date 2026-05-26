package domain

import "context"

type userContextKey string

const UserSessionKey userContextKey = "user_session"

func WithUserSession(ctx context.Context, token UserSession) context.Context {
	return context.WithValue(ctx, UserSessionKey, token)
}

func FromUserSession(ctx context.Context) (UserSession, bool) {
	value, ok := ctx.Value(UserSessionKey).(UserSession)
	return value, ok
}
