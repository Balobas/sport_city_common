package authManager

import (
	"context"
	"log"

	auth_v1 "github.com/balobas/sport_city_common/api/auth_service_v1"
	"github.com/pkg/errors"
)

// TODO: singleFlight
func (cm *ClientsAuthManager) Login(ctx context.Context) error {
	email, pwd := cm.cfg.ServiceCreds()

	resp, err := cm.client.Login(ctx, &auth_v1.LoginRequest{
		Email:    email,
		Password: pwd,
	})
	if err != nil {
		log.Printf("clientsAuthManager.Login: failed to login: %v", err)
		return errors.WithStack(err)
	}

	cm.mu.Lock()
	cm.accessJwt = resp.GetAccessJwt()
	cm.refreshJwt = resp.GetRefreshJwt()
	cm.mu.Unlock()

	return nil
}
