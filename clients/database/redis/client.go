package redisClient

import (
	"context"
	"fmt"

	"github.com/balobas/sport_city_common/logger"
	"github.com/pkg/errors"

	redisotel "github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func New(ctx context.Context, cfg Config, opts ...RedisClientOption) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.RedisPassword(),
		DB:           cfg.RedisDB(),
		Username:     cfg.RedisUser(),
		MaxRetries:   cfg.RedisMaxRetries(),
		DialTimeout:  cfg.RedisDialTimeout(),
		ReadTimeout:  cfg.RedisDialTimeout(),
		WriteTimeout: cfg.RedisTimeout(),
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	options := &clientOptions{}
	for _, apply := range opts {
		apply(options)
	}

	if options.withTracing {
		if err := redisotel.InstrumentTracing(client); err != nil {
			return nil, errors.Wrap(err, "failed to instrument redis tracing")
		}
	}

	log := logger.From(ctx)
	log.Info().Msgf("successfully connected to redis server on %s", cfg.RedisAddr())

	return &RedisClient{
		client: client,
	}, nil
}

func (c *RedisClient) SetStrWithoutExp(ctx context.Context, key string, value string) error {
	if err := c.client.Set(ctx, key, value, 0).Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *RedisClient) GetStr(ctx context.Context, key string) (string, error) {
	res, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("value for key %s not found", key)
		}
		return "", errors.WithStack(err)
	}
	return res, nil
}

func (c *RedisClient) GetStrs(ctx context.Context, keys ...string) ([]string, error) {
	r, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := make([]string, len(r))
	for i := 0; i < len(r); i++ {
		if r[i] == nil {
			res[i] = ""
			continue
		}

		val := r[i].(string)
		if val == redis.Nil.Error() {
			res[i] = ""
		} else {
			res[i] = val
		}
	}
	return res, nil
}

func (c *RedisClient) ScanAllKeys(ctx context.Context, pattern string) ([]string, error) {
	var res []string
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		res = append(res, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}

func (c *RedisClient) Delete(ctx context.Context, key string) error {
	cmd := c.client.Del(ctx, key)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (c *RedisClient) Close(ctx context.Context) error {
	return c.client.Close()
}
