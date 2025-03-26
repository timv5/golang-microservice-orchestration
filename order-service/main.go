package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"order-service/client"
	"order-service/configs"
	"order-service/handler"
	"order-service/repository"
	"order-service/route"
	"order-service/service"
	"strconv"
)

var (
	server               *gin.Engine
	OrderController      handler.OrderHandler
	OrderRouteController route.OrderRouteHandler
)

func main() {
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic("Could not initialize app")
	}

	postgresDB, err := configs.ConnectToDB(&config)
	if err != nil {
		panic("Failed to connect to DB")
	}

	walletClient := client.NewWalletClient("TODO")
	redisDatabase := initializeRedisCache(config)

	// initialize repository
	orderRepository := repository.NewOrderRepository()

	// initialize service
	redisService := service.NewRedisService(redisDatabase)
	orderService := service.NewOrderService(&config, postgresDB, orderRepository, redisService, walletClient)

	// initialize handlers
	OrderController = handler.NewOrderHandler(postgresDB, orderService, &config)
	OrderRouteController = route.NewOrderRouteHandler(OrderController)

	server = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{config.ClientOrigin}
	corsConfig.AllowCredentials = true

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	OrderRouteController.OrderRoute(router)

	log.Fatal(server.Run(":" + config.ServerPort))
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
