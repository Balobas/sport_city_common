package clientAuthService

import (
	"context"

	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
	grpcErrors "github.com/balobas/sport_city_common/grpc/errors"
	"github.com/balobas/sport_city_common/logger"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

func (c *AuthClient) Register(ctx context.Context, email string, password string) (uuid.UUID, error) {
	log := logger.From(ctx)
	log.Debug().Msgf("authClient.Register: email %s", email)

	resp, err := c.client.Register(ctx, &auth_v1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	userUid, err := uuid.FromString(resp.Uid)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(errors.New(grpcErrors.ErrBadGateway), err.Error())
	}

	return userUid, nil
}
