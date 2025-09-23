package workerOutboxPublisher

import (
	"context"
	"time"

	"github.com/balobas/sport_city_common/logger"
)

type Worker struct {
	cfg              Config
	outboxRepository OutboxRepository
	publisher        Publisher
}

func New(
	cfg Config,
	outboxRepository OutboxRepository,
	publisher Publisher,
) *Worker {
	return &Worker{
		cfg:              cfg,
		outboxRepository: outboxRepository,
		publisher:        publisher,
	}
}

func (w *Worker) Run(ctx context.Context) {
	log := logger.From(ctx).With().Fields(map[string]interface{}{
		"layer":     "worker",
		"component": "publisherWorker",
	}).Logger()
	log.Info().Msg("start publisher worker")

	msgsBatchSize := w.cfg.MqPublishMessagesBatchSize()
	timer := time.NewTimer(w.cfg.MqPublishMessagesInterval())

	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Info().Msgf("stop publisher worker. ctx done %v\n", ctx.Err())
			return
		case <-timer.C:
			select {
			case <-ctx.Done():
				log.Info().Msgf("stop publisher worker. ctx done %v\n", ctx.Err())
				return
			default:
			}

			msgs, err := w.outboxRepository.GetReadyMessagesForPublish(ctx, msgsBatchSize)
			if err != nil {
				log.Error().Err(err).Msg("error get ready messages for publish")
				timer.Reset(w.cfg.MqPublishMessagesInterval())
				break
			}

			for _, msg := range msgs {
				if err := w.publisher.Publish(ctx, msg.SubjectName, msg.Payload); err != nil {
					log.Error().Err(err).Msgf("failed to publish message %s into %s", msg.Uid, msg.SubjectName)

					msg.UpdatedAt = time.Now().UTC()
					msg.LastErrorMessage = err.Error()
				} else {
					msg.SendAt = time.Now().UTC()
					log.Info().Msgf("successfuly send message %s into %s", msg.Uid, msg.SubjectName)
				}

				if err := w.outboxRepository.UpdateMessage(ctx, msg); err != nil {
					log.Error().Msgf("failed to update message %s: %v", msg.Uid, err)
				}
			}

			timer.Reset(w.cfg.MqPublishMessagesInterval())
		}
	}
}
