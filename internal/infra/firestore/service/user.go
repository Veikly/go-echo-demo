package firestoreservice

import (
	"context"

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

func (s *User) GetUserDetailById(ctx context.Context, userId string) (*model.User, error) {
	docRef := s.client.Collection("users").Doc(userId)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, err
		}
		// 用户首次访问：从 UserSession 同步创建
		return s.syncUserFromSession(ctx, docRef)
	}
	var userDTO dto.User
	if err := docSnap.DataTo(&userDTO); err != nil {
		return nil, err
	}
	userDTO.ID = docRef.ID

	// 如果 Firebase Auth 中的 email 已变更，则同步更新
	if session, ok := domain.FromUserSession(ctx); ok && session.Email != "" && session.Email != userDTO.Email {
		_, _ = docRef.Set(ctx, map[string]any{"Email": session.Email}, firestore.MergeAll)
		userDTO.Email = session.Email
	}

	return userDTO.ToEntity(), nil
}

func (s *User) syncUserFromSession(ctx context.Context, docRef *firestore.DocumentRef) (*model.User, error) {
	session, ok := domain.FromUserSession(ctx)
	if !ok {
		// 不应发生：auth middleware 保证 UserSession 存在
		return nil, status.Error(codes.Unauthenticated, "user session not found")
	}
	userDTO := dto.NewUserFromSession(session)
	if _, err := docRef.Set(ctx, userDTO); err != nil {
		return nil, err
	}
	userDTO.ID = docRef.ID
	return userDTO.ToEntity(), nil
}

func (s *User) CompleteUserInfo(ctx context.Context, userInfo *model.User) (*model.User, error) {
	docRef := s.client.Collection("users").Doc(userInfo.ID)
	if _, err := docRef.Set(ctx, dto.ToMap(userInfo), firestore.MergeAll); err != nil {
		return nil, err
	}
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	var u dto.User
	if err := docSnap.DataTo(&u); err != nil {
		return nil, err
	}
	u.ID = docRef.ID
	return u.ToEntity(), nil
}
