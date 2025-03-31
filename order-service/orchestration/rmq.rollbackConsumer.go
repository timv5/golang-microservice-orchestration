package orchestration

import (
	"context"
	"gorm.io/gorm"
	"log"
	"order-service/configs"
	"order-service/repository"

	"github.com/streadway/amqp"
)

type RollbackConsumer struct {
	db              *gorm.DB
	channel         *amqp.Channel
	config          *configs.Config
	orderRepository *repository.OrderRepository
}

func NewRollbackConsumer(db *gorm.DB, cfg *configs.Config, ch *amqp.Channel, orderRepository *repository.OrderRepository) *RollbackConsumer {
	return &RollbackConsumer{
		db:              db,
		channel:         ch,
		config:          cfg,
		orderRepository: orderRepository,
	}
}

func (rc *RollbackConsumer) Consume(ctx context.Context) error {
	msgs, err := rc.channel.Consume(
		rc.config.RMQRollbackEventOrderQueue,
		"",
		false, // autoAck disabled so that we explicitly ack messages
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Process messages in a separate goroutine.
	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Rollback consumer channel closed")
					return
				}
				log.Printf("Rollback Consumer: Received message: %s", string(msg.Body))

				requestId := string(msg.Body)
				err = rc.processRollback(requestId)
				if err != nil {
					log.Printf("Error occured processing a rollback: %v", err)
					msg.Nack(false, false)
					continue
				}

				// Acknowledge the message
				if err := msg.Ack(false); err != nil {
					log.Printf("Failed to ack message: %v", err)
				}
			case <-ctx.Done():
				log.Println("Rollback consumer shutting down")
				return
			}
		}
	}()

	return nil
}

func (rc *RollbackConsumer) processRollback(requestId string) error {
	tx := rc.getDbConnection()

	err := rc.orderRepository.Delete(tx, requestId)
	if err != nil {
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
	}

	log.Printf("requestId %s successfully rolledback", requestId)
	return nil
}

func (rc *RollbackConsumer) getDbConnection() *gorm.DB {
	tx := rc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	return tx
}
