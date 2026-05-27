package dto

import (
	"go-echo-demo/internal/domain"
	"go-echo-demo/internal/model"
)

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
		ID:       u.ID,
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

func NewUserFromSession(session domain.UserSession) User {
	return User{
		ID:    session.UID,
		Email: session.Email,
	}
}

func ToMap(u *model.User) map[string]any {
	if u == nil {
		return nil
	}

	data := map[string]any{}

	if u.ID != "" {
		data["id"] = u.ID
	}

	if u.Username != "" {
		data["username"] = u.Username
	}

	if u.Email != "" {
		data["email"] = u.Email
	}

	if u.Age != 0 {
		data["age"] = u.Age
	}

	if u.Address.Province != "" || u.Address.City != "" || u.Address.Detail != "" {
		data["address"] = map[string]any{}

		address := data["address"].(map[string]any)

		if u.Address.Province != "" {
			address["province"] = u.Address.Province
		}

		if u.Address.City != "" {
			address["city"] = u.Address.City
		}

		if u.Address.Detail != "" {
			address["detail"] = u.Address.Detail
		}
	}

	if u.Profile.Avatar != "" || u.Profile.Bio != "" {
		data["profile"] = map[string]any{}

		profile := data["profile"].(map[string]any)

		if u.Profile.Avatar != "" {
			profile["avatar"] = u.Profile.Avatar
		}

		if u.Profile.Bio != "" {
			profile["bio"] = u.Profile.Bio
		}
	}

	return data
}
