package service

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisServiceInterface interface {
	IdempotencyValidation(key string) (bool, error)
}

type RedisService struct {
	redisClient *redis.Client
}

func NewRedisService(redisClient *redis.Client) *RedisService {
	return &RedisService{redisClient: redisClient}
}

func (red *RedisService) IdempotencyValidation(key string) (bool, error) {
	if key == "" {
		return false, nil
	}

	ctx := context.Background()
	result, err := red.redisClient.SetNX(ctx, key, "processed", 5*time.Minute).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}
