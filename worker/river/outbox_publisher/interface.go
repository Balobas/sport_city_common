package riverOutboxPublisher

import (
	"context"

	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
)

type Config interface {
	MqPublishMessagesBatchSize() int64
}

type Publisher interface {
	Publish(ctx context.Context, subjectName string, data []byte) error
}

type OutboxRepository interface {
	UpdateMessage(ctx context.Context, msg outboxEntity.Message) error
}
