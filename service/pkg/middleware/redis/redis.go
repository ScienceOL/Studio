package redis

import (
	"context"

	r "github.com/redis/go-redis/v9"
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

func CloseRedis(_ context.Context) {
	if redisClient != nil {
		redisClient.Close()
	}
}

// GetClient 获取Redis客户端实例
func GetClient() *r.Client {
	return redisClient
}
