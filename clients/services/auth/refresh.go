package clientAuthService

import (
	"context"

	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
	"github.com/balobas/sport_city_common/logger"
)

func (c *AuthClient) Refresh(ctx context.Context, refreshToken string) (accessJwt string, refreshJwt string, err error) {
	log := logger.From(ctx)
	log.Debug().Msgf("authClient.Refresh: token %s", refreshToken)

	var resp *auth_v1.JwtResponse
	resp, err = c.client.Refresh(ctx, &auth_v1.RefreshRequest{
		RefreshJwt: refreshToken,
	})
	if err != nil {
		return
	}

	accessJwt = resp.GetAccessJwt()
	refreshJwt = resp.GetRefreshJwt()
	return
}
