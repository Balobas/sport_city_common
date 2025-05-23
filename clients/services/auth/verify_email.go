package clientAuthService

import (
	"context"
	"log"

	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
)

func (c *AuthClient) VerifyEmail(ctx context.Context, token string) error {
	log.Printf("authClient.VerifyEmail: token %s", token)

	if _, err := c.client.VerifyEmail(ctx, &auth_v1.VerifyEmailRequest{
		Token: token,
	}); err != nil {
		log.Printf("failed to verify email by token %s: %v", token, err)
		return err
	}

	return nil
}
