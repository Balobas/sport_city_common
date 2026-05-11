package authInterceptor

import (
	"context"
	"strings"
	"time"

	grpcErrors "github.com/balobas/sport_city_common/grpc/errors"
	"github.com/balobas/sport_city_common/logger"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryAuthInterceptor(withoutAuthMethods map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log := logger.From(ctx)

		if _, allowWithoutAuth := withoutAuthMethods[info.FullMethod]; allowWithoutAuth {
			log.Debug().Msgf("method %s allowed without auth", info.FullMethod)
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Debug().Msgf("empty metadata for request %s", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, grpcErrors.AuthErrMsgTokenNotProvided)
		}
		accessJwtMd := md.Get("accessJwt")
		if len(accessJwtMd) == 0 {
			log.Debug().Msgf("empty accessJwt in metadata (request %s)", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, grpcErrors.AuthErrMsgTokenNotProvided)
		}

		accessJwt := accessJwtMd[0]
		if len(accessJwt) == 0 {
			log.Debug().Msgf("empty accessJwt[0] in metadata (request %s)", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, grpcErrors.AuthErrMsgTokenNotProvided)
		}

		callerInfo, err := parseAndVerifyToken(ctx, accessJwt)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		log.Info().Msgf("user %s successfully verified", callerInfo.Uid)

		log = log.With().Str("callerUid", callerInfo.Uid.String()).Logger()
		ctx = logger.ContextWithLogger(ctx, log)

		return handler(contextWithCallerInfo(ctx, callerInfo), req)
	}
}

func parseAndVerifyToken(ctx context.Context, tokenStr string) (CallerInfo, error) {
	log := logger.From(ctx)

	token, _ := jwt.Parse(tokenStr, nil)
	if token == nil {
		log.Debug().Msg("auth interceptor: failed to parse jwt token: empty token")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}
	claims, _ := token.Claims.(jwt.MapClaims)

	tokenTypeVal, ok := claims["type"]
	if !ok {
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}
	tokenType := tokenTypeVal.(string)
	if tokenType == tokenTypeSystem {
		return parseSystemToken(log, claims)
	}

	return parseUserToken(log, claims)
}

const (
	tokenTypeSystem = "system"
)

func parseUserToken(log zerolog.Logger, claims jwt.MapClaims) (CallerInfo, error) {
	userUidStr, ok := claims["user_uid"]
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: user_uid is empty")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	userUid, err := uuid.FromString(userUidStr.(string))
	if err != nil {
		log.Debug().Msg("auth interceptor: invalid jwt: user_uid is invalid")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	deviceUidStr, ok := claims["device_uid"]
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: device_uid is empty")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	deviceUid, err := uuid.FromString(deviceUidStr.(string))
	if err != nil {
		log.Debug().Msg("auth interceptor: invalid jwt: device_uid is invalid")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	roles, ok := claims["roles"]
	if !ok {
		return CallerInfo{}, errors.New("empty roles in token")
	}

	rolesStr, ok := roles.(string)
	if !ok {
		return CallerInfo{}, errors.New("invalid roles")
	}
	var rolesStrs []string
	if len(rolesStr) != 0 {
		rolesStrs = strings.Split(rolesStr, "")
	}

	expiredStr, ok := claims["expired_at"]
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: expired is empty")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	expiredAt, ok := expiredStr.(float64)
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: failed to parse expired")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	if int64(expiredAt) <= time.Now().UTC().Unix() {
		log.Debug().Msg("auth interceptor: invalid jwt: token expired")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgTokenExpired)
	}

	return CallerInfo{
		Uid:       userUid,
		DeviceUid: deviceUid,
		Roles:     rolesStrs,
		IsSystem:  false,
	}, nil
}

func parseSystemToken(log zerolog.Logger, claims jwt.MapClaims) (CallerInfo, error) {
	serviceUidStr, ok := claims["service_uid"]
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: service_uid is empty")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	serviceUid, err := uuid.FromString(serviceUidStr.(string))
	if err != nil {
		log.Debug().Msg("auth interceptor: invalid jwt: service_uid is invalid")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	deviceUidStr, ok := claims["device_uid"]
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: device_uid is empty")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	deviceUid, err := uuid.FromString(deviceUidStr.(string))
	if err != nil {
		log.Debug().Msg("auth interceptor: invalid jwt: device_uid is invalid")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	domainVal, ok := claims["domain"]
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: domain is empty")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	domain, ok := domainVal.(string)
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: domain is invalid")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	roles, ok := claims["roles"]
	if !ok {
		return CallerInfo{}, errors.New("empty roles in token")
	}

	rolesStr, ok := roles.(string)
	if !ok {
		return CallerInfo{}, errors.New("invalid roles")
	}
	var rolesStrs []string
	if len(rolesStr) != 0 {
		rolesStrs = strings.Split(rolesStr, "")
	}

	expiredStr, ok := claims["expired_at"]
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: expired is empty")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	expiredAt, ok := expiredStr.(float64)
	if !ok {
		log.Debug().Msg("auth interceptor: invalid jwt: failed to parse expired")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgInvalidToken)
	}

	if int64(expiredAt) <= time.Now().UTC().Unix() {
		log.Debug().Msg("auth interceptor: invalid jwt: token expired")
		return CallerInfo{}, errors.New(grpcErrors.AuthErrMsgTokenExpired)
	}

	return CallerInfo{
		Uid:       serviceUid,
		DeviceUid: deviceUid,
		Roles:     rolesStrs,
		IsSystem:  true,
		Domain:    domain,
	}, nil
}

type userCtxKey struct{}

func contextWithCallerInfo(ctx context.Context, info CallerInfo) context.Context {
	return context.WithValue(ctx, userCtxKey{}, info)
}

func CallerInfoFromContext(ctx context.Context) CallerInfo {
	info, ok := ctx.Value(userCtxKey{}).(CallerInfo)
	if !ok {
		return CallerInfo{}
	}
	return info
}

type CallerInfo struct {
	Uid       uuid.UUID //userUid for user token, service uid for system token
	DeviceUid uuid.UUID
	Roles     []string
	IsSystem  bool
	Domain    string
}
