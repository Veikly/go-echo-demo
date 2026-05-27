package service

import (
	"context"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain"
	"go-echo-demo/internal/infra/firestore/dto"
	"go-echo-demo/internal/model"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	client *firestore.Client
}

func NewUser(client *firestore.Client) *User {
	return &User{
		client: client,
	}
}

func upsertUser(ctx context.Context) *model.User {
	u, ok := ctx.Value(domain.UserSessionKey).(domain.UserSession)
	var user model.User
	if ok {
		user.ID = u.UID
		user.Email = u.Email
	}
	return &user
}

func (s *User) GetUserDetailById(ctx context.Context, userId string) (*model.User, error) {
	docRef := s.client.Collection("users").Doc(userId)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			newUser := upsertUser(ctx)
			s.client.Collection("users").Doc(newUser.ID).Set(ctx, newUser.ToMap())
			return newUser, constants.UserNotFound
		}
		return nil, err
	}
	var user dto.User
	if err := docSnap.DataTo(&user); err != nil {
		return nil, err
	}
	return user.ToEntity(), nil
}

func (s *User) CompleteUserInfo(ctx context.Context, userInfo *model.User) (*model.User, error) {
	docRef, _, err := s.client.Collection("users").Add(ctx, dto.ToDTO(userInfo))
	if err != nil {
		return nil, err
	}
	data, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	var u dto.User
	if err := data.DataTo(&u); err != nil {
		return nil, err
	}
	u.ID = docRef.ID
	return u.ToEntity(), nil
}
