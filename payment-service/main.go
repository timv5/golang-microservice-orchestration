package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"log"
	"payment-service/configs"
	"payment-service/handler"
	"payment-service/orchestration"
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

	rmqConn := initializeRabbitMQ(config)
	defer rmqConn.Close()
	// Initialize rollback consumer on a separate channel
	cancelRollbackConsumer := initializeRollbackConsumer(postgresDB, config, rmqConn, transactionRepository, accountRepository)
	defer cancelRollbackConsumer()

	server = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{config.ClientOrigin}
	corsConfig.AllowCredentials = true

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	PaymentRouteController.PaymentRoute(router)

	log.Fatal(server.Run(":" + config.ServerPort))
}

func initializeRabbitMQ(cfg configs.Config) *amqp.Connection {
	conn, err := amqp.Dial(cfg.RMQUrl)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Fatalf("Failed to open a RabbitMQ channel: %v", err)
	}

	_, err = ch.QueueDeclare(
		cfg.RMQExpiredEventQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		log.Fatalf("Failed to declare RabbitMQ queue: %v", err)
	}

	return conn
}

func initializeRollbackConsumer(postgresDB *gorm.DB, cfg configs.Config, conn *amqp.Connection, transactionRepository *repository.TransactionRepository, accountRepository *repository.AccountRepository) context.CancelFunc {
	rollbackChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel for rollback consumer: %v", err)
	}
	rollbackConsumer := orchestration.NewRollbackConsumer(postgresDB, &cfg, rollbackChannel, accountRepository, transactionRepository)
	consumerCtx, consumerCancel := context.WithCancel(context.Background())

	if err := rollbackConsumer.Consume(consumerCtx); err != nil {
		log.Fatalf("Failed to start rollback consumer: %v", err)
	}
	log.Println("Rollback consumer started")

	return func() {
		consumerCancel()
		rollbackChannel.Close()
	}
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
