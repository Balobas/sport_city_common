package grpcClientConn

import (
	"context"
	"strings"

	grpcErrors "github.com/balobas/sport_city_common/grpc/errors"
	"github.com/balobas/sport_city_common/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthManager interface {
	Login(ctx context.Context) error
	Register(ctx context.Context) error
	Refresh(ctx context.Context) error
	GetAccessToken() string
}

type ClientConnWithAuth struct {
	*grpc.ClientConn

	authManager AuthManager
	maxRetries  int64
}

func NewClientConnWithAuth(
	target string,
	authManager AuthManager,
	maxRetriesOnRequest int64,
	opts ...grpc.DialOption,
) (*ClientConnWithAuth, error) {
	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &ClientConnWithAuth{
		ClientConn:  cc,
		authManager: authManager,
		maxRetries:  maxRetriesOnRequest,
	}, nil
}

func (cc *ClientConnWithAuth) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return cc.invoke(ctx, method, args, reply, 0, opts...)
}

func (cc *ClientConnWithAuth) invoke(ctx context.Context, method string, args any, reply any, retries int64, opts ...grpc.CallOption) error {
	log := logger.From(ctx)
	retries++
	if retries > cc.maxRetries {
		log.Debug().Msgf("authClientConn.Invoke: max retry attempts (%d) reached", cc.maxRetries)
		return errors.New("retry attempts reached")
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"accessJwt": cc.authManager.GetAccessToken(),
	}))

	err := cc.ClientConn.Invoke(ctx, method, args, reply, opts...)
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); ok && s.Code() != codes.Unauthenticated {
		return err
	}

	log.Warn().Err(err).Msgf("authClientConn.Invoke: failed to invoke %s", method)

	if isTokenNotProvidedFromError(err) {
		log.Debug().Msg("authClientConn.Invoke: try to login")

		loginErr := cc.authManager.Login(ctx)

		if loginErr != nil {
			if isUserNotFoundFromError(loginErr) {
				log.Debug().Msg("authClientConn.Invoke: login failed, try to register")
				regErr := cc.authManager.Register(ctx)
				if regErr != nil {
					log.Warn().Err(regErr).Msg("authClientConn.Invoke: request failed, login failed, register failed")
					return err
				}

				log.Debug().Msg("authClientConn.Invoke: successfully register service")

				loginErr = cc.authManager.Login(ctx)
				if loginErr != nil {
					log.Warn().Err(loginErr).Msg("authClientConn.Invoke: request failed, login failed after register")
					return err
				}

			} else {
				log.Warn().Msgf("authClientConn.Invoke: request failed (%v), login failed %v", err, loginErr)
				return err
			}
		}
		log.Debug().Msg("authClientConn.Invoke: successfully login")

		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"accessJwt": cc.authManager.GetAccessToken(),
		}))

		return cc.invoke(ctx, method, args, reply, retries, opts...)
	}

	if isTokenInvalidOrExpiredFromError(err) {
		log.Debug().Msg("authClientConn.Invoke: token is invalid or expired, try to refresh")
		refreshErr := cc.authManager.Refresh(ctx)
		if refreshErr != nil {
			log.Warn().Msgf("authClientConn.Invoke: request failed (%v), refresh tokens failed %v", err, refreshErr)
			return err
		}
		log.Debug().Msg("authClientConn.Invoke: successfully refresh token")

		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"accessJwt": cc.authManager.GetAccessToken(),
		}))
		return cc.invoke(ctx, method, args, reply, retries, opts...)
	}

	return err
}

func isTokenNotProvidedFromError(err error) bool {
	return strings.Contains(err.Error(), grpcErrors.AuthErrMsgTokenNotProvided)
}

func isTokenInvalidOrExpiredFromError(err error) bool {
	return strings.Contains(err.Error(), grpcErrors.AuthErrMsgInvalidToken) || strings.Contains(err.Error(), grpcErrors.AuthErrMsgTokenExpired)
}

func isUserNotFoundFromError(err error) bool {
	return strings.Contains(err.Error(), userNotFound)
}

const (
	userNotFound = "user not found"
)
