package authManager

import (
	"context"

	"github.com/balobas/sport_city_common/api/auth_internal_api"
)

func (cm *ClientsAuthManager) Register(ctx context.Context) error {
	uid, pwd := cm.cfg.ServiceCreds()

	_, err := cm.client.ServiceRegister(ctx, &auth_internal_api.ServiceRegisterRequest{
		Uid:      uid,
		Name:     cm.cfg.ServiceName(),
		Domain:   cm.cfg.ServiceDomain(),
		Password: pwd,
	})
	if err != nil {
		return err
	}

	return nil
}
