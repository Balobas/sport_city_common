package authManager

import (
	"context"

	"github.com/balobas/sport_city_common/api/auth_internal_api"
)

// TODO: singleFlight
func (cm *ClientsAuthManager) Refresh(ctx context.Context) error {
	cm.mu.RLock()
	refreshJwt := cm.refreshJwt
	cm.mu.RUnlock()

	resp, err := cm.client.ServiceRefresh(ctx, &auth_internal_api.ServiceRefreshRequest{
		RefreshJwt: refreshJwt,
	})
	if err != nil {
		return err
	}

	cm.mu.Lock()
	cm.accessJwt = resp.GetAccessJwt()
	cm.refreshJwt = resp.GetRefreshJwt()
	cm.mu.Unlock()

	return nil
}
