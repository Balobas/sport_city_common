package clientAuthService

import (
	"context"

	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
	"github.com/balobas/sport_city_common/logger"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"
)

func (c *AuthClient) UpdateAuthUser(
	ctx context.Context,
	accessToken string,
	uid uuid.UUID,
	email string,
	password string,
) (accessJwt string, refreshJwt string, err error) {
	log := logger.From(ctx)
	log.Debug().Msgf("authClient.UpdateUser: uid %s", uid)

	ctx = metadata.AppendToOutgoingContext(ctx, accessJwtKey, accessToken)

	var resp *auth_v1.JwtResponse
	resp, err = c.client.UpdateUser(ctx, &auth_v1.UpdateUserRequest{
		Uid:      uid.String(),
		Email:    email,
		Password: password,
	})
	if err != nil {
		return
	}

	accessJwt = resp.GetAccessJwt()
	refreshJwt = resp.GetRefreshJwt()
	return
}
