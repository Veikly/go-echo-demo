package usecase

import (
	"context"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain"
	"go-echo-demo/internal/usecase"
	"go-echo-demo/internal/usecase/usecaseio"
)

type User struct {
	userSvc usecase.User
}

func NewUser(s usecase.User) *User {
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
	session, ok := domain.FromUserSession(ctx)
	if !ok {
		return usecaseio.CompleteUserInfoDetail{}, constants.CredentialsAbsence
	}
	userModel := usecaseio.ToModelUser(&input)
	userModel.ID = session.UID
	userModel.Email = session.Email

	res, err := u.userSvc.CompleteUserInfo(ctx, userModel)
	if err != nil {
		return usecaseio.CompleteUserInfoDetail{}, err
	}
	return usecaseio.NewCompleteUserInfoDetail(res), nil
}
