package authManager

import (
	"context"
	"log"

	auth_v1 "github.com/balobas/sport_city_common/api/auth_service_v1"
	"github.com/pkg/errors"
)

// TODO: singleFlight
func (cm *ClientsAuthManager) Refresh(ctx context.Context) error {
	cm.mu.RLock()
	refreshJwt := cm.refreshJwt
	cm.mu.RUnlock()

	resp, err := cm.client.Refresh(ctx, &auth_v1.RefreshRequest{
		RefreshJwt: refreshJwt,
	})
	if err != nil {
		log.Printf("clientsAuthManager.Refresh: failed to refresh: %v", err)
		return errors.WithStack(err)
	}

	cm.mu.Lock()
	cm.accessJwt = resp.GetAccessJwt()
	cm.refreshJwt = resp.GetRefreshJwt()
	cm.mu.Unlock()

	return nil
}
