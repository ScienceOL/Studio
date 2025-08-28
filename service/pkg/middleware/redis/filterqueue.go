package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
)

type MessageConsumer struct {
	redisClient         *redis.Client
	luaScript           *redis.Script // lua 脚本
	scheduleName        string        // 调度器随机名
	scheduleUserSet     string        // 调度器用户集合名
	retryCount          int
	maxAttempts         int   // 最大重试次数
	expireTimeoutSecond int64 // 消息过期时间，毫秒
	setExpireHours      int   // Set过期时间，小时
}

func NewMessageConsumer(redisClient *redis.Client, scheduleID string) *MessageConsumer {
	// 带时间检查和Set过期时间延长的Lua脚本
	singleScript := redis.NewScript(`
        local queue_name = KEYS[1]   -- job 队列名
        local labs_set = KEYS[2]   -- 实验用户集合
		    local retry_count = tonumber(ARGV[1]) or 3
        local expire_timeout_second = tonumber(ARGV[2]) or 3600   -- 默认60分钟
        local current_time_second = tonumber(ARGV[3])   -- 当前时间 
        local set_expire_second = tonumber(ARGV[4]) or 86400   -- 默认24小时
        local max_attempts = tonumber(ARGV[5]) or 30   -- 最大尝试次数
        
        local user_count = redis.call('SCARD', labs_set)
        if user_count == 0 then
            return nil
        end
        
        -- 延长Set的过期时间（只要有用户存在就延长）
        redis.call('EXPIRE', labs_set, set_expire_second)
        
        local attempts = 0
         
        while attempts < retry_count do
            attempts = attempts + 1
            local message = redis.call('RPOP', queue_name)
            
            if not message then
                return nil
            end
            
            local success, json_data = pcall(function()
                return cjson.decode(message)
            end)
            
            -- 如果解析失败，直接返回原消息
            if not success then
                return message
            end
            
            -- 处理 attempt_count 字段
            json_data.attempt_count = (json_data.attempt_count or 0) + 1
            
            -- 处理 enqueue_time 字段：如果没有设置则设置为当前时间
            json_data.enqueue_time = json_data.enqueue_time or current_time_second
            
            -- 检查过期：计算消息年龄（当前时间 - 入队时间）
            local message_age_second = current_time_second - json_data.enqueue_time
            if message_age_second > expire_timeout_second then
                return cjson.encode(json_data)
            end
            
            -- 检查最大尝试次数
            -- if json_data.attempt_count >= max_attempts then
            --     return cjson.encode(json_data)
            -- end
            
            -- 检查UUID
            if not json_data.uuid or json_data.uuid == "" then
                return cjson.encode(json_data)
            end
            
            -- 检查用户是否在当前Pod
            if redis.call('SISMEMBER', labs_set, json_data.uuid) == 1 then
                return cjson.encode(json_data)
            else
                redis.call('LPUSH', queue_name, cjson.encode(json_data))
            end
        end
        
        return nil
    `)

	scheduleLabKey := fmt.Sprintf("lab_websocket_uuid_%s_users", scheduleID)

	return &MessageConsumer{
		redisClient:         redisClient,
		luaScript:           singleScript,
		scheduleName:        scheduleID,
		scheduleUserSet:     scheduleLabKey,
		retryCount:          3, // 每次连续读取三条
		maxAttempts:         30,
		expireTimeoutSecond: 60 * 60, // 1 小时
		setExpireHours:      24,      // 24小时
	}
}

// 添加用户到过滤列表
func (mc *MessageConsumer) AddUser(ctx context.Context, userUUID string) error {
	err := mc.redisClient.SAdd(ctx, mc.scheduleUserSet, userUUID).Err()
	if err != nil {
		return code.RedisAddSetErr.WithMsgf("failed to add user to redis: %v", err)
	}

	// 设置Set过期时间
	mc.redisClient.Expire(ctx, mc.scheduleUserSet, time.Duration(mc.setExpireHours)*time.Hour)
	return nil
}

// 从过滤列表移除用户
func (mc *MessageConsumer) RemoveUser(ctx context.Context, userUUID string) error {
	if err := mc.redisClient.SRem(ctx, mc.scheduleUserSet, userUUID).Err(); err != nil {
		return code.RedisRemoveSetErr.WithMsgf("User %s removed from schedule %s filter", userUUID, mc.scheduleName)
	}

	return nil
}

// 删除整个用户集合
func (mc *MessageConsumer) DeleteUserSet(ctx context.Context) error {
	err := mc.redisClient.Del(ctx, mc.scheduleUserSet).Err()
	if err != nil {
		logger.Errorf(ctx, "DeleteUserSet err: %+v", err)
		return code.RedisRemoveSetErr.WithMsgf("failed to delete user set from redis: %+v", err)
	}
	return nil
}

// 检查用户是否存在
func (mc *MessageConsumer) HasUser(ctx context.Context, userUUID string) (bool, error) {
	exists, err := mc.redisClient.SIsMember(ctx, mc.scheduleUserSet, userUUID).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// 消费消息 - 修正版本
func (mc *MessageConsumer) consumeMessage(ctx context.Context, queueName string) ([]byte, error) {
	// 获取当前时间戳（毫秒）
	currentTimeSecond := time.Now().Unix()
	setExpireSeconds := int64(mc.setExpireHours * 60 * 60) // 转换为秒

	cmd := mc.luaScript.Run(ctx, mc.redisClient,
		[]string{queueName, mc.scheduleUserSet},
		mc.retryCount, mc.expireTimeoutSecond, currentTimeSecond, setExpireSeconds, mc.maxAttempts)

	result, err := cmd.Result()
	if err != nil {
		// 检查是否是 redis.Nil 错误（表示脚本返回了 nil）
		if errors.Is(err, redis.Nil) || err.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, code.RedisLuaScriptErr.WithErr(err)
	}

	if result == nil {
		return nil, nil
	}

	messageStr, ok := result.(string)
	if !ok {
		return nil, code.RedisLuaRetErr
	}

	return []byte(messageStr), nil
}

func (mc *MessageConsumer) Message(ctx context.Context, queueName string, messageHandler func([]byte)) {
	message, err := mc.consumeMessage(ctx, queueName)
	if err != nil {
		logger.Errorf(ctx, "Error consuming message: %v", err)
		return
	}

	// 如果没有消息，这是正常情况，不需要日志
	if len(message) != 0 && messageHandler != nil {
		messageHandler(message)
	}
}

// 设置最大尝试次数
func (mc *MessageConsumer) SetMaxAttempts(attempts int) {
	mc.maxAttempts = attempts
}

// 设置消息过期时间（分钟）
func (mc *MessageConsumer) SetMessageExpireTimeout(minutes int) {
	mc.expireTimeoutSecond = int64(minutes * 60 * 1000)
}

// 设置Set过期时间（小时）
func (mc *MessageConsumer) SetUserSetExpireTimeout(hours int) {
	mc.setExpireHours = hours
}

// 清理资源
func (mc *MessageConsumer) Cleanup(ctx context.Context) error {
	return mc.DeleteUserSet(ctx)
}
