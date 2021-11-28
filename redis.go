package grok

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// NewRedisConnection ...
func NewRedisConnection(connectionString string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: connectionString,
	})

	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		logrus.WithError(err).Panic("Error pinging Redis")
	}

	return client
}
