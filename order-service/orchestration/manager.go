package orchestration

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"order-service/configs"
	"time"
)

type ManagerInterface interface {
	Start(orchestrationId string) error
	End(orchestrationId string) error
	Rollback(orchestrationId string) error
}

type Manager struct {
	redisClient *redis.Client
	rmqProducer *RMQProducer
	config      *configs.Config
}

func NewOrchestrationManager(redisClient *redis.Client, rmqProducer *RMQProducer, config *configs.Config) *Manager {
	return &Manager{
		redisClient: redisClient,
		rmqProducer: rmqProducer,
		config:      config,
	}
}

func (or *Manager) Start(orchestrationId string) error {
	entity := or.build(orchestrationId, InProgress)
	data, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	err = or.redisClient.HSet(context.Background(), or.config.OrchestrationMapName, orchestrationId, data).Err()
	if err != nil {
		return err
	}
	return nil
}

func (or *Manager) End(orchestrationId string) error {
	err := or.redisClient.HDel(context.Background(), or.config.OrchestrationMapName, orchestrationId).Err()
	if err != nil {
		return err
	}
	return nil
}

func (or *Manager) Rollback(orchestrationId string) error {
	// publish to rmq so that orchestration will pick it up and started rollback
	err := or.rmqProducer.Produce(orchestrationId)
	if err != nil {
		return err
	}
	return nil
}

func (or *Manager) build(orchestrationId string, status string) Model {
	orchestrationModel := Model{
		UUID:           orchestrationId,
		Status:         status,
		ExpirationTime: time.Now().UnixMilli() + or.config.OrchestrationExpirationTimeSeconds*1000,
	}
	return orchestrationModel
}
