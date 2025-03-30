package orchestration

import (
	"errors"
	"order-service/configs"

	"github.com/streadway/amqp"
)

type RMQProducerInterface interface {
	Produce(orchestrationId string) error
}

type RMQProducer struct {
	config  *configs.Config
	channel *amqp.Channel
}

func NewRMQProducer(cfg *configs.Config, channel *amqp.Channel) *RMQProducer {
	return &RMQProducer{
		config:  cfg,
		channel: channel,
	}
}

func (prod *RMQProducer) Produce(orchestrationId string) error {
	if prod.channel == nil {
		return errors.New("rabbit channel not initialized")
	}
	return prod.channel.Publish(
		"",
		prod.config.RMQExpiredEventQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(orchestrationId),
		},
	)
}
