package natsClient

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	mqClient "github.com/balobas/sport_city_common/clients/mq"
	"github.com/balobas/sport_city_common/logger"
	"github.com/balobas/sport_city_common/tracer"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
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
	log := logger.Logger()

	conn, err := nats.Connect(
		cfg.NatsUrl(), nats.Name(cfg.NatsClientName()),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Info().Msg("nats has been recconected")
		}),
		nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
			log.Info().Msgf("nats error handler: error occured: sub %s : %v", s.Subject, err)
		}),
		nats.DisconnectHandler(func(c *nats.Conn) {
			log.Info().Msg("nats disconnect")
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Info().Msg("nats closed")
		}),
		nats.ConnectHandler(func(c *nats.Conn) {
			log.Info().Msgf("nats successfully connected to %s", c.ConnectedAddr())
		}),
		nats.RetryOnFailedConnect(true),
	)
	if err != nil {
		log.Debug().Err(err).Msgf("failed to connect to nats (url: %s)", cfg.NatsUrl())
		return nil, err
	}

	return &NatsClientPubSub{conn: conn}, nil
}

func (nc *NatsClientPubSub) Publish(ctx context.Context, subj string, data []byte) error {
	return nc.conn.Publish(subj, data)
}

func (nc *NatsClientPubSub) Subscribe(ctx context.Context, handlers map[string]map[string]mqClient.MqMsgHandler) error {
	log := logger.Logger()

	h := handlers[mqClient.PubsubKey]
	for subject, handler := range h {

		_, err := nc.conn.Subscribe(subject, convertToNatsMsgHandler(ctx, handler))
		if err != nil {
			log.Debug().Msgf("failed to subscribe on subject %s", subject)
			return errors.WithStack(err)
		}
		log.Info().Msgf("successfully subscribed on %s", subject)
	}
	return nil
}

func (nc *NatsClientPubSub) Close(ctx context.Context) error {
	nc.conn.Close()
	return nil
}

func convertToNatsMsgHandler(ctx context.Context, handler mqClient.MqMsgHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		log := logger.Logger()

		if err := handler(ctx, msg.Data); err != nil {
			log.Error().Err(err).Msgf("failed to handle message %s", msg.Data)
			nackWithLog(msg)
			return
		}

		if err := msg.Ack(); err != nil {
			log.Error().Err(err).Msgf("failed to ack message %s", msg.Data)
		}
	}
}

func nackWithLog(msg *nats.Msg) {
	if err := msg.Nak(); err != nil {
		log := logger.Logger()
		log.Error().Err(err).Msgf("failed to nack msg %s", msg.Data)
	}
}

type NatsClientJetStream struct {
	cfg           Config
	conn          *nats.Conn
	js            jetstream.JetStream
	consumersCtxs []jetstream.ConsumeContext

	opts NatsClientJsOpts

	wg        *sync.WaitGroup
	connected chan struct{}
}

type NatsClientJetStreamOption interface {
	Apply(*NatsClientJsOpts)
}

type NatsClientJsOpts struct {
	withoutNackOnErrors bool
	withTrace           bool
}

type natsJsOptionWithoutNackOnErrors bool

func (no *natsJsOptionWithoutNackOnErrors) Apply(opt *NatsClientJsOpts) {
	opt.withoutNackOnErrors = bool(*no)
}

func WithoutNackOnErrors() NatsClientJetStreamOption {
	n := natsJsOptionWithoutNackOnErrors(true)
	return &n
}

type natsJsOptionWithTrace bool

func (nt *natsJsOptionWithTrace) Apply(opt *NatsClientJsOpts) {
	opt.withTrace = bool(*nt)
}

func WithTrace() NatsClientJetStreamOption {
	n := natsJsOptionWithTrace(true)
	return &n
}

func NewJs(ctx context.Context, cfg Config, opts ...NatsClientJetStreamOption) (mqClient.MqClient, error) {
	log := logger.From(ctx)
	connectedChan := make(chan struct{})

	conn, err := nats.Connect(
		cfg.NatsUrl(), nats.Name(cfg.NatsClientName()),
		nats.ReconnectHandler(func(c *nats.Conn) {
			select {
			case connectedChan <- struct{}{}:
			default:
			}
			log.Info().Msg("nats has been recconected")
		}),
		nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
			log.Info().Msgf("nats error handler: error occured: sub %s", s.Subject)
		}),
		nats.DisconnectHandler(func(c *nats.Conn) {
			log.Info().Msg("nats disconnect")
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Info().Msg("nats closed")
		}),
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(c *nats.Conn) {

			select {
			case connectedChan <- struct{}{}:
			default:
			}
			log.Info().Msgf("nats successfully connected to %s", c.ConnectedAddr())
		}),
		nats.RetryOnFailedConnect(true),
	)
	if err != nil {
		log.Debug().Err(err).Msgf("failed to connect to nats (url: %s)", cfg.NatsUrl())
		return nil, err
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		log.Debug().Err(err).Msg("failed to init nats jet stream")
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
		opt.Apply(&nc.opts)
	}

	return nc, nil
}

func (nc *NatsClientJetStream) Publish(ctx context.Context, subj string, data []byte) error {
	log := logger.From(ctx)
	ack, err := nc.js.Publish(ctx, subj, data)
	if err != nil {
		log.Debug().Msgf("failed to publish message into subject %s", subj)
		return errors.WithStack(err)
	}
	log.Info().Msgf("ack info: stream %s, domain %s, duplicate %t, sequence %d", ack.Stream, ack.Domain, ack.Duplicate, ack.Sequence)
	return nil
}

func (nc *NatsClientJetStream) Subscribe(ctx context.Context, handlersStreams map[string]map[string]mqClient.MqMsgHandler) (subErr error) {
	log := logger.From(ctx)
	failedStreams := map[string]map[string]mqClient.MqMsgHandler{}

	defer func() {
		if subErr == nil {
			nc.resubscribeOnFailedStreams(ctx, failedStreams)
		}
	}()

	for streamName, handlers := range handlersStreams {

		stream, err := nc.js.Stream(ctx, streamName)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to get stream %s", streamName)
			failedStreams[streamName] = handlers
			continue
		}

		for subject, handler := range handlers {

			subjectParts := strings.Split(subject, ".")
			if len(subjectParts) != 2 {
				log.Debug().Err(err).Msgf("invalid message subject %s", subject)
				return errors.Errorf("invalid message subject %s", subject)
			}

			consumerName := fmt.Sprintf("%s_%s_consumer", nc.cfg.ServiceName(), strings.Split(subject, ".")[1])
			consumer, err := stream.Consumer(ctx, consumerName)
			if err != nil {
				log.Warn().Err(err).Msgf("failed to get consumer %s on stream %s subject %s", consumerName, streamName, subject)
				addSubjectToFailedStreams(streamName, subject, handler, failedStreams)
				continue
			}

			consumerCtx, err := consumer.Consume(nc.convertToNatsJsMsgHandler(ctx, handler))
			if err != nil {
				log.Warn().Err(err).Msgf("failed to init consumer %s on stream %s subject %s", consumerName, streamName, subject)
				addSubjectToFailedStreams(streamName, subject, handler, failedStreams)
				continue
			}
			log.Info().Msgf("successfully init consumer %s on stream %s subject %s", consumerName, streamName, subject)

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
	log := logger.From(ctx)

	nc.wg.Add(1)
	go func() {
		defer nc.wg.Done()

		select {
		case <-ctx.Done():
			log.Info().Msg("stop attempts to init failed consumers, ctx done")
			return
		case <-nc.connected:
			select {
			case <-ctx.Done():
				log.Info().Msg("stop attempts to init failed consumers, ctx done")
				return
			default:
			}

			nc.Subscribe(ctx, failedStreams)
		case <-time.After(1 * time.Minute):
			select {
			case <-ctx.Done():
				log.Info().Msg("stop attempts to init failed consumers, ctx done")
				return
			default:
			}

			nc.Subscribe(ctx, failedStreams)
		}
	}()
}

func (nc *NatsClientJetStream) Close(ctx context.Context) error {
	log := logger.From(ctx)

	for _, consumerCtx := range nc.consumersCtxs {
		consumerCtx.Stop()
	}
	nc.conn.Close()
	nc.wg.Wait()
	close(nc.connected)

	log.Info().Msg("nats js client closed successfully")
	return nil
}

type msgWithTraceId struct {
	TraceId string `json:"traceId"`
}

func (nc *NatsClientJetStream) convertToNatsJsMsgHandler(ctx context.Context, handler mqClient.MqMsgHandler) jetstream.MessageHandler {
	var (
		parentStructName string
		fnName           string
	)
	fnOrMethodRealName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()

	parentStructParts := strings.Split(fnOrMethodRealName, "(*")
	if len(parentStructParts) != 0 {
		parentStructName = strings.Split(parentStructParts[1], ")")[0]
	}

	fnNameParts := strings.Split(fnOrMethodRealName, ".")
	if len(fnNameParts) != 0 {
		fnName = strings.Split(fnNameParts[len(fnNameParts)-1], "-")[0]
	}

	return func(msg jetstream.Msg) {

		log := logger.From(ctx).With().Fields(map[string]interface{}{
			"layer":     "handlers",
			"layerType": "mq",
			"component": parentStructName,
			"method":    fnName,
		}).Logger()

		if nc.opts.withTrace {
			var msgTraceInfo msgWithTraceId
			if err := json.Unmarshal(msg.Data(), &msgTraceInfo); err != nil {
				log.Warn().Str("error", err.Error()).Msg("failed to unmarshal traceID from message")
			} else {
				traceId, err := trace.TraceIDFromHex(msgTraceInfo.TraceId)
				if err != nil {
					log.Error().Err(err).Msg("invalid trace id")
				} else {
					spanContext := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID: traceId,
					})
					ctx = trace.ContextWithSpanContext(ctx, spanContext)
				}
			}
			var span trace.Span
			ctx, span = tracer.FromCtx(ctx).Start(ctx, fmt.Sprintf("%s.%s", parentStructName, fnName))
			defer span.End()
		}

		if err := handler(ctx, msg.Data()); err != nil {
			log.Error().Err(err).Msgf("failed to handle message %s", msg.Data())
			if nc.opts.withoutNackOnErrors {
				return
			}

			nackJsMsgWithLog(ctx, msg)
			return
		}

		if err := msg.Ack(); err != nil {
			log.Error().Err(err).Msgf("failed to ack message %s", msg.Data())
		}
	}
}

func nackJsMsgWithLog(ctx context.Context, msg jetstream.Msg) {
	log := logger.From(ctx)
	// default delay, TODO: refactor
	if err := msg.NakWithDelay(1 * time.Second); err != nil {
		log.Error().Err(err).Msgf("failed to nack msg %s", msg.Data())
	}
}
