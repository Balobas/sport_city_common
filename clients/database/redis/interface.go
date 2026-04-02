package redisClient

import "time"

type Config interface {
	RedisAddr() string
	RedisPassword() string
	RedisUser() string
	RedisDB() int
	RedisMaxRetries() int
	RedisDialTimeout() time.Duration
	RedisTimeout() time.Duration
}
