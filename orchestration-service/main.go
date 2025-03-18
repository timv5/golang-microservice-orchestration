package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"orchestration-service/configs"
	"strconv"
)

func main() {
	// set configs
	//config, err := configs.LoadConfig(".")
	//if err != nil {
	//	panic("Could not initialize app")
	//}

	//redisDatabase := initializeRedisCache(config)
}

func initializeRedisCache(config configs.Config) *redis.Client {
	redisDb, err := strconv.Atoi(config.RedisDb)
	if err != nil {
		panic("Could not initialize app, error converting redis db configs")
	}
	redisDatabase := redis.NewClient(&redis.Options{
		Addr:     config.RedisHost + ":" + config.RedisPort,
		Password: "",
		DB:       redisDb,
	})

	// Ping the Redis server to check the connection
	ctx := context.Background()
	pong, err := redisDatabase.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Redis ping response:", pong)

	return redisDatabase
}
