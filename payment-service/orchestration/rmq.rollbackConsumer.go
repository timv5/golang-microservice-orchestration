package orchestration

import (
	"context"
	"gorm.io/gorm"
	"log"
	"payment-service/configs"
	"payment-service/repository"

	"github.com/streadway/amqp"
)

type RollbackConsumer struct {
	db                    *gorm.DB
	channel               *amqp.Channel
	config                *configs.Config
	accountRepository     *repository.AccountRepository
	transactionRepository *repository.TransactionRepository
}

func NewRollbackConsumer(
	db *gorm.DB,
	cfg *configs.Config,
	ch *amqp.Channel,
	accountRepository *repository.AccountRepository,
	transactionRepository *repository.TransactionRepository) *RollbackConsumer {
	return &RollbackConsumer{
		db:                    db,
		channel:               ch,
		config:                cfg,
		accountRepository:     accountRepository,
		transactionRepository: transactionRepository,
	}
}

func (rc *RollbackConsumer) Consume(ctx context.Context) error {
	msgs, err := rc.channel.Consume(
		rc.config.RMQRollbackEventPaymentQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

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
	transaction, err := rc.transactionRepository.Delete(tx, requestId)
	if err != nil {
		return err
	}

	err = rc.accountRepository.AddAmount(tx, transaction.AccountId, transaction.Amount)
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
