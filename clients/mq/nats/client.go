package natsClient

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	mqClient "github.com/balobas/sport_city_common/clients/mq"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
)

type Config interface {
	NatsUrl() string
	NatsClientName() string
	ServiceName() string
}

type NatsClientPubSub struct {
	conn *nats.Conn
}

func NewPubSub(cfg Config) (mqClient.MqClient, error) {
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

	return &NatsClientPubSub{conn: conn}, nil
}

func (nc *NatsClientPubSub) Publish(ctx context.Context, subj string, data []byte) error {
	return nc.conn.Publish(subj, data)
}

func (nc *NatsClientPubSub) Subscribe(ctx context.Context, handlers map[string]map[string]mqClient.MqMsgHandler) error {

	h := handlers[mqClient.PubsubKey]
	for subject, handler := range h {

		_, err := nc.conn.Subscribe(subject, convertToNatsMsgHandler(ctx, handler))
		if err != nil {
			log.Printf("failed to subscribe on subject %s: %v", subject, err)
			return errors.WithStack(err)
		}
		log.Printf("successfully subscribed on %s", subject)
	}
	return nil
}

func (nc *NatsClientPubSub) Close(ctx context.Context) error {
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

type NatsClientJetStream struct {
	cfg           Config
	conn          *nats.Conn
	js            jetstream.JetStream
	consumersCtxs []jetstream.ConsumeContext

	opts *NatsClientJsOpts

	wg        *sync.WaitGroup
	connected chan struct{}
}

type NatsClientJetStreamOption interface {
	Apply(*NatsClientJsOpts)
}

type NatsClientJsOpts struct {
	withoutNackOnErrors bool
}

type natsJsOptionWithoutNackOnErrors bool

func (no *natsJsOptionWithoutNackOnErrors) Apply(opt *NatsClientJsOpts) {
	opt.withoutNackOnErrors = bool(*no)
}

func WithoutNackOnErrors() NatsClientJetStreamOption {
	n := natsJsOptionWithoutNackOnErrors(true)
	return &n
}

func NewJs(ctx context.Context, cfg Config, opts ...NatsClientJetStreamOption) (mqClient.MqClient, error) {
	connectedChan := make(chan struct{})

	conn, err := nats.Connect(
		cfg.NatsUrl(), nats.Name(cfg.NatsClientName()),
		nats.ReconnectHandler(func(c *nats.Conn) {
			select {
			case connectedChan <- struct{}{}:
			default:
			}
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
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(c *nats.Conn) {

			select {
			case connectedChan <- struct{}{}:
			default:
			}
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

	nc := &NatsClientJetStream{
		cfg:       cfg,
		conn:      conn,
		js:        js,
		wg:        &sync.WaitGroup{},
		connected: connectedChan,
	}

	for _, opt := range opts {
		opt.Apply(nc.opts)
	}

	return nc, nil
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

func (nc *NatsClientJetStream) Subscribe(ctx context.Context, handlersStreams map[string]map[string]mqClient.MqMsgHandler) (subErr error) {
	failedStreams := map[string]map[string]mqClient.MqMsgHandler{}

	defer func() {
		if subErr == nil {
			nc.resubscribeOnFailedStreams(ctx, failedStreams)
		}
	}()

	for streamName, handlers := range handlersStreams {

		stream, err := nc.js.Stream(ctx, streamName)
		if err != nil {
			log.Printf("failed to get stream %s: %v", streamName, err)
			failedStreams[streamName] = handlers
			continue
		}

		for subject, handler := range handlers {

			subjectParts := strings.Split(subject, ".")
			if len(subjectParts) != 2 {
				log.Printf("invalid message subject %s", subject)
				return errors.Errorf("invalid message subject %s", subject)
			}

			consumerName := fmt.Sprintf("%s_%s_consumer", nc.cfg.ServiceName(), strings.Split(subject, ".")[1])
			consumer, err := stream.Consumer(ctx, consumerName)
			if err != nil {
				log.Printf("failed to get consumer %s on stream %s subject %s: %v", consumerName, streamName, subject, err)
				addSubjectToFailedStreams(streamName, subject, handler, failedStreams)
				continue
			}

			consumerCtx, err := consumer.Consume(nc.convertToNatsJsMsgHandler(ctx, handler))
			if err != nil {
				log.Printf("failed to init consumer %s on stream %s subject %s: %v", consumerName, streamName, subject, err)
				addSubjectToFailedStreams(streamName, subject, handler, failedStreams)
				continue
			}
			log.Printf("successfully init consumer %s on stream %s subject %s", consumerName, streamName, subject)

			nc.consumersCtxs = append(nc.consumersCtxs, consumerCtx)
		}
	}

	return nil
}

func addSubjectToFailedStreams(
	streamName string,
	subject string,
	handler mqClient.MqMsgHandler,
	failedStreams map[string]map[string]mqClient.MqMsgHandler,
) {
	failedStream, ok := failedStreams[streamName]
	if !ok {
		failedStreams[streamName] = map[string]mqClient.MqMsgHandler{
			subject: handler,
		}
	} else {
		failedStream[subject] = handler
	}
}

func (nc *NatsClientJetStream) resubscribeOnFailedStreams(ctx context.Context, failedStreams map[string]map[string]mqClient.MqMsgHandler) {
	if len(failedStreams) == 0 {
		return
	}

	nc.wg.Add(1)
	go func() {
		defer nc.wg.Done()

		select {
		case <-ctx.Done():
			log.Printf("stop attempts to init failed consumers")
			return
		case <-nc.connected:
			select {
			case <-ctx.Done():
				log.Printf("stop attempts to init failed consumers")
				return
			default:
			}

			nc.Subscribe(ctx, failedStreams)
		case <-time.After(1 * time.Minute):
			select {
			case <-ctx.Done():
				log.Printf("stop attempts to init failed consumers")
				return
			default:
			}

			nc.Subscribe(ctx, failedStreams)
		}
	}()
}

func (nc *NatsClientJetStream) Close(ctx context.Context) error {
	for _, consumerCtx := range nc.consumersCtxs {
		consumerCtx.Stop()
	}
	nc.conn.Close()
	nc.wg.Wait()
	close(nc.connected)
	log.Printf("nats js client closed successfully")
	return nil
}

func (nc *NatsClientJetStream) convertToNatsJsMsgHandler(ctx context.Context, handler mqClient.MqMsgHandler) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		if err := handler(ctx, msg.Data()); err != nil {
			log.Printf("failed to handle message %s: %v", msg.Data(), err)
			if nc.opts.withoutNackOnErrors {
				return
			}

			nackJsMsgWithLog(msg)
			return
		}

		if err := msg.Ack(); err != nil {
			log.Printf("failed to ack message %s: %v", msg.Data(), err)
		}
	}
}

func nackJsMsgWithLog(msg jetstream.Msg) {
	// default delay, TODO: refactor
	if err := msg.NakWithDelay(1 * time.Second); err != nil {
		log.Printf("failed to nack msg %s: %v", msg.Data(), err)
	}
}
