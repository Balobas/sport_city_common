package workerOutboxPublisher

import (
	"context"
	"time"

	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
)

type Config interface {
	MqPublishMessagesInterval() time.Duration
	MqPublishMessagesBatchSize() int64
}

type Publisher interface {
	Publish(ctx context.Context, subjectName string, data []byte) error
}

type OutboxRepository interface {
	GetReadyMessagesForPublish(ctx context.Context, batchSize int64) ([]outboxEntity.Message, error)
	UpdateMessage(ctx context.Context, msg outboxEntity.Message) error
}
