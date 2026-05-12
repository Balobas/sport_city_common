package authManager

import (
	"context"

	"github.com/balobas/sport_city_common/api/auth_internal_api"
	"github.com/pkg/errors"
)

// TODO: singleFlight
func (cm *ClientsAuthManager) Login(ctx context.Context) error {
	uid, pwd := cm.cfg.ServiceCreds()

	resp, err := cm.client.ServiceLogin(ctx, &auth_internal_api.ServiceLoginRequest{
		Uid:      uid,
		Password: pwd,
		Device: &auth_internal_api.Device{
			Uid:  cm.cfg.ServiceDeviceUid().String(),
			Name: cm.cfg.ServiceDeviceName(),
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}

	cm.mu.Lock()
	cm.accessJwt = resp.GetAccessJwt()
	cm.refreshJwt = resp.GetRefreshJwt()
	cm.mu.Unlock()

	return nil
}
