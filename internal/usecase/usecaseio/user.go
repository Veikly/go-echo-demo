package usecaseio

import (
	"go-echo-demo/internal/model"
)

type Address struct {
	Province string
	City     string
	Detail   string
}

type Profile struct {
	Avatar string
	Bio    string
}

type UserDetailOutput struct {
	Username string
	Email    string
	Age      int
	Address  Address
	Profile  Profile
}

type CompleteUserInfoDetail struct {
	ID       string
	Username string
	Email    string
	Age      int
	Address  Address
	Profile  Profile
}

func NewUserDetailOutput(u *model.User) UserDetailOutput {
	return UserDetailOutput{
		Username: u.Username,
		Email:    u.Email,
		Age:      u.Age,
		Address:  Address(u.Address),
		Profile:  Profile(u.Profile),
	}
}

func ToModelUser(info *CompleteUserInfoDetail) *model.User {
	return &model.User{
		Username: info.Username,
		Email:    info.Email,
		Age:      info.Age,
		Address:  model.Address(info.Address),
		Profile:  model.Profile(info.Profile),
	}
}

func NewCompleteUserInfoDetail(u *model.User) CompleteUserInfoDetail {
	return CompleteUserInfoDetail{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Age:      u.Age,
		Address:  Address(u.Address),
		Profile:  Profile(u.Profile),
	}
}
