package authInterceptor

import (
	"context"
	"log"
	"strings"
	"time"

	grpcErrors "github.com/balobas/sport_city_common/grpc/errors"
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
			return nil, status.Error(codes.Unauthenticated, grpcErrors.AuthErrMsgTokenNotProvided)
		}
		accessJwtMd := md.Get("accessJwt")
		if len(accessJwtMd) == 0 {
			log.Printf("empty accessJwt in metadata")
			return nil, status.Error(codes.Unauthenticated, grpcErrors.AuthErrMsgTokenNotProvided)
		}

		accessJwt := accessJwtMd[0]
		if len(accessJwt) == 0 {
			log.Printf("empty accessJwt[0] in metadata")
			return nil, status.Error(codes.Unauthenticated, grpcErrors.AuthErrMsgTokenNotProvided)
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
		return AuthUserInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	userUidStr, ok := claims["user_uid"]
	if !ok {
		log.Printf("auth interceptor: invalid jwt: user_uid is empty")
		return AuthUserInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	userUid, err := uuid.FromString(userUidStr.(string))
	if err != nil {
		log.Printf("auth interceptor: invalid jwt: user_uid is invalid")
		return AuthUserInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	roles, ok := claims["roles"]
	if !ok {
		return AuthUserInfo{}, errors.New("empty roles in token")
	}

	rolesStr, ok := roles.(string)
	if !ok {
		return AuthUserInfo{}, errors.New("invalid roles")
	}
	var rolesStrs []string
	if len(rolesStr) != 0 {
		rolesStrs = strings.Split(rolesStr, "")
	}

	expiredStr, ok := claims["expired_at"]
	if !ok {
		log.Printf("auth interceptor: invalid jwt: expired is empty")
		return AuthUserInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	// ??????
	expiredAt, ok := expiredStr.(float64)
	if !ok {
		log.Printf("auth interceptor: invalid jwt: failed to parse expired")
		return AuthUserInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	if int64(expiredAt) <= time.Now().UTC().Unix() {
		log.Printf("auth interceptor: invalid jwt: token expired")
		return AuthUserInfo{}, errors.New(grpcErrors.AuthErrMsgTokenExpired)
	}

	userInfo := AuthUserInfo{
		UserUid: userUid,
		Roles:   rolesStrs,
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
	UserUid uuid.UUID
	Roles   []string
}
