package useCaseMediaOutbox

import (
	"context"

	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
)

type OutboxRepository interface {
	CreateMessage(ctx context.Context, message outboxEntity.Message) error
	GetReadyMessagesForPublish(ctx context.Context, batchSize int64) ([]outboxEntity.Message, error)
	UpdateMessage(ctx context.Context, message outboxEntity.Message) error
}

type Config interface {
	MediaUsageMessageSubject() string
}
