package natsClient

import (
	"context"
	"log"
	mqClient "github.com/balobas/sport_city_common/clients/mq"
)

type Config interface {
	NatsUrl() string
	NatsClientName() string
}

type NatsClient struct {
	conn *nats.Conn
}

func New(cfg Config) (mqClient.MqClient, error) {
	conn, err := nats.Connect(cfg.NatsUrl(), nats.Name(cfg.NatsClientName()))
	if err != nil {
		log.Printf("failed to connect to nats (url: %s): %v", cfg.NatsUrl(), err)
	}

	return &NatsClient{conn: conn}, nil
}

func (nc *NatsClient) Publish(subj string, data []byte) error {
	return nc.conn.Publish(subj, data)
}

func (nc *NatsClient) Subscribe(ctx context.Context, handlers map[string]mqClient.MqMsgHandler) error {
	for subject, handler := range handlers {

		_, err := nc.conn.Subscribe(subject, convertToNatsMsgHandler(ctx, handler))
		if err != nil {
			log.Printf("failed to subscribe on subject %s: %v", subject, err)
			return errors.WithStack(err)
		}
		log.Printf("successfully subscribed on %s", subject)
	}
	return nil
}

func (nc *NatsClient) Close(ctx context.Context) error {
	nc.conn.Close()
	return nil
}

func convertToNatsMsgHandler(ctx context.Context, handler mqClient.MqMsgHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		if err := handler(ctx, msg.Data); err != nil {
			log.Printf("failed to handle message %s: %v", msg.Data, err)
			nackWithLog(msg)
			return
		}

		if err := msg.Ack(); err != nil {
			log.Printf("failed to ack message %s: %v", msg.Data, err)
		}
	}
}

func nackWithLog(msg *nats.Msg) {
	if err := msg.Nak(); err != nil {
		log.Printf("failed to nack msg %s: %v", msg.Data, err)
	}
}
