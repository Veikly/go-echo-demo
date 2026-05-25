package repository

import (
	"context"
	"go-echo-demo/internal/model"
)

type User interface {
	GetUserDetailById(ctx context.Context, userId string) (*model.User, error)
	CompleteUserInfo(ctx context.Context, userInfo *model.User) (*model.User, error)
}
