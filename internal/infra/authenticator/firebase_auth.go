package authenticator

import (
	"context"
	"errors"
	"go-echo-demo/internal/domain"

	"firebase.google.com/go/v4/auth"
	"go.uber.org/zap"
)

type FirebaseAuthenticator struct {
	authClient *auth.Client
}

func NewFirebaseAuthenticator(authClient *auth.Client) *FirebaseAuthenticator {
	return &FirebaseAuthenticator{
		authClient: authClient,
	}
}

func (firebaseAuthenticator *FirebaseAuthenticator) Authenticate(ctx context.Context, tokenString string) (*domain.UserSession, error) {
	// 1. 验证 Token
	token, err := firebaseAuthenticator.authClient.VerifyIDToken(ctx, tokenString)
	if err != nil {
		zap.L().Error("Firebase Token 验证失败", zap.Error(err))
		return nil, err
	}

	// auth.Token的Claims字段被打上了 json:"-" 无法打印出来 要获取实际的认证数据 需要从token.Claims获取
	zap.L().Info("Firebase Token 验证成功，当前 Claims 结构", zap.Any("claims", token.Claims))

	// 2.检查邮件地址是否已验证 如果没有 中断这里的认证流程 提示先完成邮箱地址验证
	emailVerified, ok := token.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return nil, errors.New("please check your email address")
	}

	var u domain.UserSession

	// 2. 从 "user_id" 获取 UID
	uid, _ := token.Claims["user_id"].(string)

	// 3. 获取 Email
	email, _ := token.Claims["email"].(string)

	// 4. 校验
	if uid != "" && email != "" {
		u.Email = email
		u.UID = uid
		zap.L().Info("用户认证成功", zap.String("uid", uid), zap.String("email", email))
		return &u, nil
	}

	zap.L().Error("Token 有效但缺少关键字段", zap.String("extracted_uid", uid), zap.String("extracted_email", email))
	return nil, errors.New("token claims check error: missing uid or email")
}
