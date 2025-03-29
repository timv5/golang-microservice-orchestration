package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"orchestration-service/config"
	"orchestration-service/constants"
	"orchestration-service/dto"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

type App struct {
	redisClient *redis.Client
	rabbitCh    *amqp.Channel
	config      config.Config
}

// periodically checks for expired orchestrations in Redis.
func (a *App) expiredOrchestrationJob(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(a.config.StaleJobSchedulePeriodMilliseconds) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.processExpiredOrchestrations(ctx)
		case <-ctx.Done():
			log.Println("Expired orchestration job shutting down")
			return
		}
	}
}

// scans the Redis hash for expired records and updates them.
func (a *App) processExpiredOrchestrations(ctx context.Context) {
	now := time.Now().UnixMilli()
	records, err := a.redisClient.HGetAll(ctx, a.config.OrchestrationMapName).Result()
	if err != nil {
		log.Println("Error fetching records from Redis:", err)
		return
	}
	for id, val := range records {
		var entity dto.OrchestrationEntity
		if err := json.Unmarshal([]byte(val), &entity); err != nil {
			log.Println("Error unmarshaling record", id, ":", err)
			continue
		}

		if entity.ExpirationTime < now {
			newEntity := dto.OrchestrationEntity{
				UUID:           entity.UUID,
				Status:         constants.StatusRollback,
				ExpirationTime: time.Now().UnixMilli() + a.config.OrchestrationExpirationTimeSeconds*1000,
			}
			if err := a.updateIfMatch(ctx, id, entity, newEntity); err != nil {
				log.Println("Failed to update orchestration", id, ":", err)
			} else {
				log.Println("Found expired orchestration", id, "set to ROLLBACK")
				if err := a.publishRMQEvent(id, a.config.RMQExpiredEventQueue); err != nil {
					log.Println("Failed to publish rollback message for", id, ":", err)
				}
			}
		}
	}
}

// atomically replaces a record in Redis if the current value matches what is expected.
func (a *App) updateIfMatch(ctx context.Context, key string, oldEntity, newEntity dto.OrchestrationEntity) error {
	return a.redisClient.Watch(ctx, func(tx *redis.Tx) error {
		currentVal, err := tx.HGet(ctx, a.config.OrchestrationMapName, key).Result()
		if err != nil {
			return err
		}
		var currentEntity dto.OrchestrationEntity
		if err := json.Unmarshal([]byte(currentVal), &currentEntity); err != nil {
			return err
		}
		if currentEntity != oldEntity {
			return errors.New("record changed")
		}
		newBytes, err := json.Marshal(newEntity)
		if err != nil {
			return err
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, a.config.OrchestrationMapName, key, newBytes)
			return nil
		})
		return err
	}, a.config.OrchestrationMapName)
}

func (a *App) publishRMQEvent(orchestrationId string, rmqName string) error {
	return a.rabbitCh.Publish(
		"",
		rmqName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(orchestrationId),
		},
	)
}

// orchestrator/application will publish on this queue once it will need to rollback
// performs the rollback steps for a given orchestration ID.
// It updates the record to a rollback-in-progress state, publishes the rollback event to RabbitMQ (later orchestrator/application will took over and rollback on it's side)
// and finally removes the record from Redis.
func (a *App) processRollback(ctx context.Context, orchestrationId string) error {
	val, err := a.redisClient.HGet(ctx, a.config.OrchestrationMapName, orchestrationId).Result()
	if err == redis.Nil {
		log.Println("Orchestration", orchestrationId, "not found")
		return nil
	} else if err != nil {
		return err
	}

	var entity dto.OrchestrationEntity
	if err := json.Unmarshal([]byte(val), &entity); err != nil {
		return err
	}
	if entity.Status != constants.StatusRollback {
		log.Println("Orchestration", orchestrationId, "status is not ROLLBACK, skipping")
		return nil
	}
	newEntity := dto.OrchestrationEntity{
		UUID:           entity.UUID,
		Status:         constants.StatusRollbackInProgress,
		ExpirationTime: time.Now().UnixMilli() + a.config.OrchestrationExpirationTimeSeconds*1000,
	}
	if err := a.updateIfMatch(ctx, orchestrationId, entity, newEntity); err != nil {
		log.Println("Failed to update orchestration to IN_PROGRESS for", orchestrationId, ":", err)
		return err
	}
	log.Println("Processing rollback for orchestration", orchestrationId)

	if err := a.publishRMQEvent(orchestrationId, a.config.RMQRollbackEventQueue); err != nil {
		log.Println("Failed to publish rollback event for", orchestrationId, ":", err)
		return err
	}
	log.Println("Published rollback event for orchestration", orchestrationId)

	// Remove the orchestration record after successful event publication.
	if err := a.redisClient.HDel(ctx, a.config.OrchestrationMapName, orchestrationId).Err(); err != nil {
		return err
	}
	log.Println("Rollback processed for orchestration", orchestrationId)
	return nil
}

// listens to the RabbitMQ queue and processes rollback messages.
func (a *App) rollbackConsumer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	msgs, err := a.rabbitCh.Consume(
		a.config.RMQExpiredEventQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to register a consumer:", err)
		return
	}

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("RabbitMQ channel closed")
				return
			}
			go func(m amqp.Delivery) {
				err := a.processRollback(ctx, string(m.Body))
				if err != nil {
					log.Println("Error processing rollback for", string(m.Body), ":", err)
					m.Nack(false, true)
				} else {
					m.Ack(false)
				}
			}(msg)
		case <-ctx.Done():
			log.Println("Rollback consumer shutting down")
			return
		}
	}
}

func initRedis(cfg config.Config) (*redis.Client, error) {
	addr := cfg.RedisHost + ":" + cfg.RedisPort
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   cfg.RedisDB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func initRabbitMQ(cfg config.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(cfg.RMQUrl)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
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
		return nil, nil, err
	}
	_, err = ch.QueueDeclare(
		cfg.RMQRollbackEventQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

func initApp(cfg config.Config) (*App, *amqp.Connection, error) {
	redisClient, err := initRedis(cfg)
	if err != nil {
		return nil, nil, err
	}

	rabbitConn, rabbitCh, err := initRabbitMQ(cfg)
	if err != nil {
		return nil, nil, err
	}

	app := &App{
		redisClient: redisClient,
		rabbitCh:    rabbitCh,
		config:      cfg,
	}

	return app, rabbitConn, nil
}

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	app, rabbitConn, err := initApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	// Setup context and wait group for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// Start expired orchestration job.
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.expiredOrchestrationJob(ctx)
	}()

	// Start rollback consumer.
	wg.Add(1)
	go app.rollbackConsumer(ctx, &wg)

	// Listen for shutdown signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Service started. Press Ctrl+C to shutdown.")
	<-sigChan
	log.Println("Shutdown signal received.")
	cancel()
	wg.Wait()
	log.Println("Service shutdown gracefully.")
}
