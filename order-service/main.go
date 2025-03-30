package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"log"
	"order-service/client"
	"order-service/configs"
	"order-service/handler"
	"order-service/orchestration"
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
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		panic("Could not initialize app")
	}

	postgresDB, err := configs.ConnectToDB(&cfg)
	if err != nil {
		panic("Failed to connect to DB")
	}

	paymentClient := client.NewPaymentClient(cfg.PaymentClientBaseUrl)
	redisDatabase := initializeRedisCache(cfg)

	rmqConn, rmqChannel := initializeRabbitMQ(cfg)
	defer rmqConn.Close()
	rmqProducer := orchestration.NewRMQProducer(&cfg, rmqChannel)
	orchestrationManager := orchestration.NewOrchestrationManager(redisDatabase, rmqProducer, &cfg)

	orderRepository := repository.NewOrderRepository()
	productRepository := repository.NewProductRepository()

	redisService := service.NewRedisService(redisDatabase)
	orderService := service.NewOrderService(&cfg, postgresDB, orderRepository, redisService, paymentClient, productRepository, orchestrationManager)

	OrderController = handler.NewOrderHandler(postgresDB, orderService, &cfg)
	OrderRouteController = route.NewOrderRouteHandler(OrderController)

	// Initialize rollback consumer on a separate channel
	cancelRollbackConsumer := initializeRollbackConsumer(cfg, rmqConn)
	defer cancelRollbackConsumer()

	server = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{cfg.ClientOrigin}
	corsConfig.AllowCredentials = true
	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	OrderRouteController.OrderRoute(router)

	log.Fatal(server.Run(":" + cfg.ServerPort))
}

func initializeRollbackConsumer(cfg configs.Config, conn *amqp.Connection) context.CancelFunc {
	rollbackChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel for rollback consumer: %v", err)
	}
	rollbackConsumer := orchestration.NewRollbackConsumer(&cfg, rollbackChannel)
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

func initializeRedisCache(cfg configs.Config) *redis.Client {
	redisDb, err := strconv.Atoi(cfg.RedisDb)
	if err != nil {
		panic("Could not initialize app, error converting redis db config")
	}
	redisDatabase := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: "",
		DB:       redisDb,
	})

	ctx := context.Background()
	pong, err := redisDatabase.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Redis ping response:", pong)

	return redisDatabase
}

func initializeRabbitMQ(cfg configs.Config) (*amqp.Connection, *amqp.Channel) {
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

	return conn, ch
}
