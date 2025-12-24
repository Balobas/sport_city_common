package commonConfig

import (
	"context"
	"os"
	"strconv"

	"github.com/balobas/sport_city_common/logger"
)

func parseIntFromEnvWithDefaultOnErr(envName string, defaultValue int) int {
	log := logger.From(context.Background())
	val, err := strconv.Atoi(os.Getenv(envName))
	if err != nil {
		log.Warn().Msgf("failed to parse env %s: %v", envName, err)
		return defaultValue
	}
	return val
}
