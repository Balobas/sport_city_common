package commonConfig

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/balobas/sport_city_common/logger"
)

type RiverConfig struct {
	queues            map[string]int // key: queueName, value: maxWorkers
	queueNames        []string
	maxAttempts       int
	maxWorkers        int
	nextRetry         Duration
	jobTimeout        Duration
	fetchCooldown     Duration
	fetchPollInterval Duration
}

const (
	RiverEnvPrefix            = "RIVER"
	RiverEnvQueueNames        = "RIVER_QUEUE_NAMES"
	RiverEnvMaxAttempts       = "RIVER_MAX_ATTEMPTS"
	RiverEnvMaxWorkers        = "RIVER_MAX_WORKERS"
	RiverEnvNextRetry         = "RIVER_NEXT_RETRY"
	RiverEnvJobTimeout        = "RIVER_JOB_TIMEOUT"
	RiverEnvFetchCooldown     = "RIVER_FETCH_COOLDOWN"
	RiverEnvFetchPollInterval = "RIVER_FETCH_POLL_INTERVAL"
	RiverEnvMaxWorkersSuffix  = "MAX_WORKERS"
)

const (
	riverMaxAttemptsDefault = 5
	riverMaxWorkersDefault  = 10
)

func ParseRiverConfig() *RiverConfig {
	cfg := &RiverConfig{}

	cfg.queueNames = strings.Split(os.Getenv(RiverEnvQueueNames), ",")

	cfg.maxAttempts = parseIntFromEnvWithDefaultOnErr(RiverEnvMaxAttempts, riverMaxAttemptsDefault)
	if cfg.maxAttempts < 1 {
		cfg.maxAttempts = riverMaxAttemptsDefault
	}

	cfg.maxWorkers = parseIntFromEnvWithDefaultOnErr(RiverEnvMaxWorkers, riverMaxWorkersDefault)
	if cfg.maxWorkers < 1 {
		cfg.maxWorkers = riverMaxWorkersDefault
	}

	cfg.queues = parseQueuesMaxWorkersByQueueNames(cfg.queueNames, cfg.maxWorkers)

	cfg.nextRetry.ParseFromEnvWithDefaultOnErr(RiverEnvNextRetry, Duration{30 * time.Second})
	cfg.jobTimeout.ParseFromEnvWithDefaultOnErr(RiverEnvJobTimeout, Duration{30 * time.Second})
	cfg.fetchCooldown.ParseFromEnvWithDefaultOnErr(RiverEnvFetchCooldown, Duration{5 * time.Second})
	cfg.fetchPollInterval.ParseFromEnvWithDefaultOnErr(RiverEnvFetchPollInterval, Duration{10 * time.Second})

	return cfg
}

func parseQueuesMaxWorkersByQueueNames(names []string, defaultMaxWorkers int) map[string]int {
	res := make(map[string]int, len(names))
	log := logger.From(context.Background())
	for _, name := range names {
		envName := RiverEnvPrefix + "_" + strings.ToUpper(name) + "_" + RiverEnvMaxWorkersSuffix
		val := os.Getenv(envName)
		maxWorkers, err := strconv.Atoi(val)
		if err != nil {
			log.Warn().Msgf("failed to parse env %s: %v", envName, err)
		}

		if maxWorkers < 1 {
			maxWorkers = defaultMaxWorkers
		}

		res[name] = maxWorkers
	}
	return res
}

func (rc *RiverConfig) Queues() map[string]int {
	return rc.queues
}

func (rc *RiverConfig) QueueNames() []string {
	return rc.queueNames
}

func (rc *RiverConfig) MaxAttempts() int {
	return rc.maxAttempts
}

func (rc *RiverConfig) MaxWorkers() int {
	return rc.maxWorkers
}

func (rc *RiverConfig) NextRetry() time.Duration {
	return rc.nextRetry.Duration
}

func (rc *RiverConfig) JobTimeout() time.Duration {
	return rc.jobTimeout.Duration
}

func (rc *RiverConfig) FetchCooldown() time.Duration {
	return rc.fetchCooldown.Duration
}

func (rc *RiverConfig) FetchPollInterval() time.Duration {
	return rc.fetchPollInterval.Duration
}
