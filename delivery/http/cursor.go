package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain"
	"strings"
)

func EncodeCursor(c *domain.PageCursor, secret []byte) (string, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(secret)
	sig := mac.Sum(nil) // 计算最终的签名

	// 返回最终结果
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawStdEncoding.EncodeToString(sig), nil
}

func DecodeCursor(raw string, secret []byte) (*domain.PageCursor, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return nil, constants.InvalidCursor
	}

	// 解码sig & payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, constants.InvalidCursor
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, constants.InvalidCursor
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(sig)
	exceptedSig := mac.Sum(nil)

	if !hmac.Equal(sig, exceptedSig) {
		return nil, constants.InvalidCursor
	}

	var cursor domain.PageCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return nil, constants.InvalidCursor
	}

	return &cursor, nil
}
