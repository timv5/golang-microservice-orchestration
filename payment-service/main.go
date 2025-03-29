package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"payment-service/configs"
	"payment-service/handler"
	"payment-service/repository"
	"payment-service/route"
	"payment-service/service"
	"strconv"
)

var (
	server                 *gin.Engine
	PaymentController      handler.PaymentHandler
	PaymentRouteController route.PaymentRouteHandler
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

	redisDatabase := initializeRedisCache(config)
	accountRepository := repository.NewAccountRepository()
	transactionRepository := repository.NewTransactionRepository()

	redisService := service.NewRedisService(redisDatabase)
	paymentService := service.NewPaymentService(postgresDB, accountRepository, transactionRepository, redisService)

	// initialize handlers
	PaymentController = handler.NewPaymentHandler(postgresDB, paymentService, &config)
	PaymentRouteController = route.NewPaymentRouteHandler(PaymentController)

	server = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{config.ClientOrigin}
	corsConfig.AllowCredentials = true

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	PaymentRouteController.PaymentRoute(router)

	log.Fatal(server.Run(":" + config.ServerPort))
}

func initializeRedisCache(config configs.Config) *redis.Client {
	redisDb, err := strconv.Atoi(config.RedisDb)
	if err != nil {
		panic("Could not initialize app, error converting redis db config")
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
