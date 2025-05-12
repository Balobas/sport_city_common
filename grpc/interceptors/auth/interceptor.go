package authInterceptor

import (
	"context"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryAuthInterceptor(withoutAuthMethods map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		if _, allowWithoutAuth := withoutAuthMethods[info.FullMethod]; allowWithoutAuth {
			log.Printf("method %s allowed without auth", info.FullMethod)
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Printf("empty metadata")
			return nil, status.Error(codes.Unauthenticated, "token not provided")
		}
		accessJwtMd := md.Get("accessJwt")
		if len(accessJwtMd) == 0 {
			log.Printf("empty accessJwt in metadata")
			return nil, status.Error(codes.Unauthenticated, "token not provided")
		}

		accessJwt := accessJwtMd[0]
		if len(accessJwt) == 0 {
			log.Printf("empty accessJwt[0] in metadata")
			return nil, status.Error(codes.Unauthenticated, "token not provided")
		}

		authUserInfo, err := parseAndVerifyToken(accessJwt)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		log.Printf("user %s successfully verified", authUserInfo.UserUid)

		return handler(contextWithUserInfo(ctx, authUserInfo), req)
	}
}

func parseAndVerifyToken(tokenStr string) (AuthUserInfo, error) {
	token, _ := jwt.Parse(tokenStr, nil)
	if token == nil {
		log.Printf("auth interceptor: failed to parse jwt token: empty token")
		return AuthUserInfo{}, errors.New("invalid token")
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	userUidStr, ok := claims["user_uid"]
	if !ok {
		log.Printf("auth interceptor: invalid jwt: user_uid is empty")
		return AuthUserInfo{}, errors.New("invalid token")
	}

	userUid, err := uuid.FromString(userUidStr.(string))
	if err != nil {
		log.Printf("auth interceptor: invalid jwt: user_uid is invalid")
		return AuthUserInfo{}, errors.New("invalid token")
	}

	userRole, ok := claims["role"]
	if !ok {
		log.Printf("auth interceptor: invalid jwt: role is empty")
		return AuthUserInfo{}, errors.New("invalid token")
	}

	userPermissions, ok := claims["permissions"]
	if !ok {
		log.Printf("auth interceptor: invalid jwt: permissions is empty")
		return AuthUserInfo{}, errors.New("invalid token")
	}

	expiredStr, ok := claims["expired_at"]
	if !ok {
		log.Printf("auth interceptor: invalid jwt: expired is empty")
		return AuthUserInfo{}, errors.New("invalid token")
	}

	// ??????
	expiredAt, ok := expiredStr.(float64)
	if !ok {
		log.Printf("auth interceptor: invalid jwt: failed to parse expired")
		return AuthUserInfo{}, errors.Errorf("invalid token: failed to parse expired")
	}

	if int64(expiredAt) <= time.Now().UTC().Unix() {
		log.Printf("auth interceptor: invalid jwt: token expired")
		return AuthUserInfo{}, errors.New("token expired")
	}

	userInfo := AuthUserInfo{
		UserUid:     userUid,
		Role:        userRole.(string),
		Permissions: userPermissions.(string),
	}
	return userInfo, nil
}

type userCtxKey struct{}

func contextWithUserInfo(ctx context.Context, info AuthUserInfo) context.Context {
	return context.WithValue(ctx, userCtxKey{}, info)
}

func UserInfoFromContext(ctx context.Context) AuthUserInfo {
	info, ok := ctx.Value(userCtxKey{}).(AuthUserInfo)
	if !ok {
		return AuthUserInfo{}
	}
	return info
}

type AuthUserInfo struct {
	UserUid     uuid.UUID
	Role        string
	Permissions string
}
