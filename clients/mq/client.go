package mqClient

import "context"

type MqMsgHandler func(ctx context.Context, msgPayload []byte) error

type MqClient interface {
	Publish(ctx context.Context, subj string, data []byte) error
	Subscribe(ctx context.Context, handlers map[string]map[string]MqMsgHandler) error
	Close(ctx context.Context) error
}

const PubsubKey = "pubsub"
