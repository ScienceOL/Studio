package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Init 初始化redis连接
func initRedis(conf *Redis) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
	client.AddHook(redisotel.NewTracingHook(redisotel.WithAttributes(semconv.NetPeerNameKey.String(addr))))
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
