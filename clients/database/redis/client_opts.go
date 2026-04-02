package redisClient

type clientOptions struct {
	withTracing bool
}

type RedisClientOption func(opts *clientOptions)

func WithTracing() RedisClientOption {
	return func(opts *clientOptions) {
		opts.withTracing = true
	}
}
