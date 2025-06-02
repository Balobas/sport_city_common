package clientAuthService

import (
	"context"

	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
	"github.com/balobas/sport_city_common/logger"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"
)

func (c *AuthClient) Logout(ctx context.Context, accessJwt string, uid uuid.UUID) error {
	log := logger.From(ctx)
	log.Debug().Msgf("authClient.Logout: uid %s", uid)

	ctx = metadata.AppendToOutgoingContext(ctx, accessJwtKey, accessJwt)

	if _, err := c.client.Logout(ctx, &auth_v1.LogoutRequest{
		Uid: uid.String(),
	}); err != nil {
		return err
	}

	return nil
}
