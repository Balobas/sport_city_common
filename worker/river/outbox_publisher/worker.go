package riverOutboxPublisher

import (
	"context"
	"time"

	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
	"github.com/balobas/sport_city_common/logger"
	"github.com/riverqueue/river"
)

type Worker struct {
	nextRetry time.Duration
	timeout   time.Duration

	publisher Publisher
}

func New(
	publisher Publisher,
) *Worker {
	return &Worker{
		publisher: publisher,
	}
}

type Args struct {
	Message outboxEntity.Message
}

func (arg Args) Kind() string {
	return "outbox_messages"
}

func (w *Worker) NextRetry(*river.Job[Args]) time.Time {
	return time.Now().Add(w.nextRetry)
}

func (w *Worker) Timeout(*river.Job[Args]) time.Duration {
	return w.timeout
}

func (w *Worker) Work(ctx context.Context, job *river.Job[Args]) error {
	log := logger.From(ctx).With().Fields(map[string]interface{}{
		"layer":     "worker",
		"component": "publisherWorker",
	}).Logger()

	msg := job.Args.Message

	if err := w.publisher.Publish(ctx, msg.SubjectName, msg.Payload); err != nil {
		log.Error().Err(err).Msgf("failed to publish message %s into %s", msg.Uid, msg.SubjectName)
		return err
	}
	return nil
}
