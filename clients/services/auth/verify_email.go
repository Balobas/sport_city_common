package clientAuthService

import (
	"context"

	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
	"github.com/balobas/sport_city_common/logger"
)

func (c *AuthClient) VerifyEmail(ctx context.Context, token string) error {
	log := logger.From(ctx)
	log.Debug().Msgf("authClient.VerifyEmail: token %s", token)

	if _, err := c.client.VerifyEmail(ctx, &auth_v1.VerifyEmailRequest{
		Token: token,
	}); err != nil {
		return err
	}

	return nil
}
