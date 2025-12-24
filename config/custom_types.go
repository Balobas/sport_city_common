package commonConfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/balobas/sport_city_common/logger"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if len(str) == 0 {
		return nil
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	if duration < 0 {
		return fmt.Errorf("duration is negative")
	}
	*d = Duration{duration}
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	if d.Duration < 0 {
		return []byte("null"), nil
	}
	str := d.Duration.String()
	data, err := json.Marshal(str)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Duration) ParseFromEnvWithDefaultOnErr(envName string, defaultValue Duration) {
	log := logger.From(context.Background())

	err := json.Unmarshal([]byte(os.Getenv(envName)), d)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to parse env %s: %v", envName, err)
	}
	if d.Seconds() == 0 {
		*d = defaultValue
	}
}
