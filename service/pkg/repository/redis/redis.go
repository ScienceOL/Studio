package redis

import (
	"context"

	r "github.com/go-redis/redis/v8"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

var redisClient *r.Client

func InitRedis(ctx context.Context, conf *Redis) {
	var err error
	redisClient, err = initRedis(conf)
	if err != nil {
		logger.Fatalf(ctx, "init redis fail err: %+v", err)
	}
}

func CloseRedis(ctx context.Context) {
	if redisClient != nil {
		redisClient.Close()
	}
}
