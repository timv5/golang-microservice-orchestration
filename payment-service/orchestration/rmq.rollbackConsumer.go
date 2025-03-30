package orchestration

import (
	"context"
	"log"
	"payment-service/configs"

	"github.com/streadway/amqp"
)

type RollbackConsumer struct {
	channel *amqp.Channel
	config  *configs.Config
}

func NewRollbackConsumer(cfg *configs.Config, ch *amqp.Channel) *RollbackConsumer {
	return &RollbackConsumer{
		channel: ch,
		config:  cfg,
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
