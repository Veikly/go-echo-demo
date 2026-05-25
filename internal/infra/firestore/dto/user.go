package dto

import "go-echo-demo/internal/model"

type Address struct {
	Province string `firestore:"province"`
	City     string `firestore:"city"`
	Detail   string `firestore:"detail"`
}

type Profile struct {
	Avatar string `firestore:"avatar"`
	Bio    string `firestore:"bio"`
}

type User struct {
	ID       string  `firestore:"-"`
	Username string  `firestore:"username"`
	Email    string  `firestore:"email"`
	Age      int     `firestore:"age"`
	Address  Address `firestore:"address"`
	Profile  Profile `firestore:"profile"`
}

func (u *User) ToEntity() *model.User {
	return &model.User{
		Username: u.Username,
		Email:    u.Email,
		Age:      u.Age,
		Address:  model.Address(u.Address),
		Profile:  model.Profile(u.Profile),
	}
}

func ToDTO(u *model.User) User {
	return User{
		Username: u.Username,
		Email:    u.Email,
		Age:      u.Age,
		Profile:  Profile(u.Profile),
		Address:  Address(u.Address),
	}
}
