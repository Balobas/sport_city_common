package httpUtils

import (
	"context"
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

type ctxKeyAuthUserInfo struct{}

type AuthUserInfo struct {
	Uid        uuid.UUID
	Email      string
	Roles      []string
	IsVerified bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	JwtToken   string
}

func ContextWithAuthUserInfo(ctx context.Context, info AuthUserInfo) context.Context {
	return context.WithValue(ctx, ctxKeyAuthUserInfo{}, info)
}

func AuthUserInfoFromContext(ctx context.Context) (AuthUserInfo, error) {
	userInfoFromCtx := ctx.Value(ctxKeyAuthUserInfo{})
	userInfo, ok := userInfoFromCtx.(AuthUserInfo)
	if !ok {
		return AuthUserInfo{}, errors.New("failed to get user")
	}

	return userInfo, nil
}
