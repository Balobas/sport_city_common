package clientAuthService

import (
	"context"

	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
	"github.com/balobas/sport_city_common/logger"
)

func (c *AuthClient) Login(ctx context.Context, email string, password string) (accessJwt string, refreshJwt string, err error) {
	log := logger.From(ctx)
	log.Debug().Msgf("authClient.Login: email %s", email)

	var resp *auth_v1.JwtResponse
	resp, err = c.client.Login(ctx, &auth_v1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return
	}

	accessJwt = resp.AccessJwt
	refreshJwt = resp.RefreshJwt
	return
}
