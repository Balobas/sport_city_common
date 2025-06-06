package authManager

import (
	"context"

	auth_v1 "github.com/balobas/sport_city_common/api/auth_service_v1"
	"github.com/pkg/errors"
)

func (cm *ClientsAuthManager) Register(ctx context.Context) error {
	email, pwd := cm.cfg.ServiceCreds()

	_, err := cm.client.Register(ctx, &auth_v1.RegisterRequest{
		Email:    email,
		Password: pwd,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
