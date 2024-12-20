package natsClient

import (
	"context"
	"log"

	mqClient "github.com/balobas/sport_city_common/clients/mq"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
)

type NatsClientJetStream struct {
	conn          *nats.Conn
	js            jetstream.JetStream
	consumersCtxs []jetstream.ConsumeContext
}

func NewJs(ctx context.Context, cfg Config) (mqClient.MqClient, error) {
	conn, err := nats.Connect(
		cfg.NatsUrl(), nats.Name(cfg.NatsClientName()),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Printf("nats has been recconected")
		}),
		nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
			log.Printf("nats error handler: error occured: sub %s : %v", s.Subject, err)
		}),
		nats.DisconnectHandler(func(c *nats.Conn) {
			log.Printf("nats disconnect")
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Printf("nats closed")
		}),
		nats.ConnectHandler(func(c *nats.Conn) {
			log.Printf("nats successfully connected to %s", c.ConnectedAddr())
		}),
		nats.RetryOnFailedConnect(true),
	)
	if err != nil {
		log.Printf("failed to connect to nats (url: %s): %v", cfg.NatsUrl(), err)
		return nil, err
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		log.Printf("failed to init nats jet stream: %v", err)
		return nil, errors.WithStack(err)
	}

	// js.Stream(ctx, cfg.UsersStreamName())

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     cfg.UsersStreamName(),
		Subjects: []string{},
		Storage:  jetstream.FileStorage,
	})
	if err != nil {
		conn.Close()
		log.Printf("failed to createOrUpdate nats stream %s: %v", cfg.UsersStreamName(), err)
		return nil, errors.WithStack(err)
	}

	return &NatsClientJetStream{
		conn: conn,
		js:   js,
	}, nil
}

func (nc *NatsClientJetStream) Publish(ctx context.Context, subj string, data []byte) error {
	ack, err := nc.js.Publish(ctx, subj, data)
	if err != nil {
		log.Printf("failed to publish message into subject %s: %v", subj, err)
		return errors.WithStack(err)
	}
	log.Printf("ack info: stream %s, domain %s, duplicate %t, sequence %d", ack.Stream, ack.Domain, ack.Duplicate, ack.Sequence)
	return nil
}

func (nc *NatsClientJetStream) Subscribe(ctx context.Context, handlersStreams map[string]map[string]mqClient.MqMsgHandler) error {

	for streamName, handlers := range handlersStreams {
		for subject, handler := range handlers {

			stream, err := nc.js.Stream(ctx, streamName)
			if err != nil {
				log.Printf("failed to get stream %s: %v", streamName, err)
				return errors.WithStack(err)
			}

			consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
				Name:          subject + "_consumer",
				DeliverPolicy: jetstream.DeliverAllPolicy,
				AckPolicy:     jetstream.AckExplicitPolicy,
				FilterSubject: subject,
			})
			if err != nil {
				log.Printf("failed to create consumer on stream %s subject %s: %v", streamName, subject, err)
				return errors.WithStack(err)
			}

			consumerCtx, err := consumer.Consume(convertToNatsJsMsgHandler(ctx, handler))
			if err != nil {
				log.Printf("failed to init consumer on stream %s subject %s: %v", streamName, subject, err)
				return errors.WithStack(err)
			}
			log.Printf("successfully init consumer on stream %s subject %s", streamName, subject)

			nc.consumersCtxs = append(nc.consumersCtxs, consumerCtx)
		}
	}

	return nil
}

func (nc *NatsClientJetStream) Close(ctx context.Context) error {
	for _, consumerCtx := range nc.consumersCtxs {
		consumerCtx.Stop()
	}
	nc.conn.Close()
	return nil
}

func convertToNatsJsMsgHandler(ctx context.Context, handler mqClient.MqMsgHandler) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		if err := handler(ctx, msg.Data()); err != nil {
			log.Printf("failed to handle message %s: %v", msg.Data(), err)
			nackJsMsgWithLog(msg)
			return
		}

		if err := msg.Ack(); err != nil {
			log.Printf("failed to ack message %s: %v", msg.Data(), err)
		}
	}
}

func nackJsMsgWithLog(msg jetstream.Msg) {
	if err := msg.Nak(); err != nil {
		log.Printf("failed to nack msg %s: %v", msg.Data(), err)
	}
}
