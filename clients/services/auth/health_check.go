package clientAuthService

import (
	"context"
	"log"

	authClientEntity "github.com/balobas/sport_city_common/clients/services/auth/entity"
)

func (c *AuthClient) HealthCheck(ctx context.Context) authClientEntity.AuthServiceHealthCheck {
	log.Printf("authClient.HealthCheck")

	resp, err := c.client.HealthCheck(ctx, nil)
	if err != nil {
		return authClientEntity.AuthServiceHealthCheck{
			Status:  authClientEntity.HealthStatusNotAvailable,
			Message: err.Error(),
		}
	}
	return authClientEntity.AuthServiceHealthCheck{
		Status:    resp.GetStatus(),
		GitTag:    resp.GetGitTag(),
		GitBranch: resp.GetGitBranch(),
		UpTime:    resp.GetUpTime().AsTime(),
	}
}
