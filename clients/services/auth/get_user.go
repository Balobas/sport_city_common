package clientAuthService

import (
	"context"

	authClientEntity "github.com/balobas/sport_city_common/clients/services/auth/entity"
	"github.com/balobas/sport_city_common/clients/services/auth/proto_gen/auth_v1"
	"github.com/balobas/sport_city_common/logger"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"
)

func (c *AuthClient) GetAuthUser(ctx context.Context, accessJwt string, uid uuid.UUID, email string) (authClientEntity.AuthUser, error) {
	log := logger.From(ctx)
	log.Debug().Msgf("authClient.GetUser: uid '%s' email '%s'", uid, email)

	ctx = metadata.AppendToOutgoingContext(ctx, accessJwtKey, accessJwt)

	resp, err := c.client.GetUser(ctx, &auth_v1.GetUserRequest{
		Uid:   uid.String(),
		Email: email,
	})
	if err != nil {
		return authClientEntity.AuthUser{}, err
	}

	return authClientEntity.AuthUser{
		Uid:         uuid.FromStringOrNil(resp.GetUid()),
		Email:       resp.GetEmail(),
		Role:        authClientEntity.UserRole(resp.GetRole().String()),
		Permissions: authClientEntity.PermissionsFromStrings(resp.GetPermissions()),
		CreatedAt:   resp.GetCreatedAt().AsTime(),
		UpdatedAt:   resp.GetUpdatedAt().AsTime(),
	}, nil
}
