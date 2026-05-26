package usecase

import (
	"context"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/usecase/repository"
	"go-echo-demo/internal/usecase/usecaseio"
)

type User struct {
	userSvc repository.User
}

func NewUser(s repository.User) *User {
	return &User{
		userSvc: s,
	}
}

func (u *User) GetMyDetail(ctx context.Context, userId string) (usecaseio.UserDetailOutput, error) {
	if userId == "" {
		return usecaseio.UserDetailOutput{}, constants.InvalidInputParam
	}
	res, err := u.userSvc.GetUserDetailById(ctx, userId)
	if err != nil {
		return usecaseio.UserDetailOutput{}, err
	}
	//entity转UserDetailOutput
	return usecaseio.NewUserDetailOutput(res), nil
}

func (u *User) CompleteUserInfo(ctx context.Context, input usecaseio.CompleteUserInfoDetail) (usecaseio.CompleteUserInfoDetail, error) {
	if input == (usecaseio.CompleteUserInfoDetail{}) {
		return usecaseio.CompleteUserInfoDetail{}, constants.InvalidInputParam
	}
	res, err := u.userSvc.CompleteUserInfo(ctx, usecaseio.ToModelUser(&input))
	if err != nil {
		return usecaseio.CompleteUserInfoDetail{}, err
	}
	return usecaseio.NewCompleteUserInfoDetail(res), nil
}
