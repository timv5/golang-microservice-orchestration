package service

import (
	"errors"
	"gorm.io/gorm"
	"order-service/client"
	"order-service/configs"
	"order-service/dto/request"
	"order-service/dto/response"
	"order-service/repository"
)

type OrderServiceInterface interface {
	Create(request request.OrderRequest) (response.OrderResponse, error)
}

type OrderService struct {
	conf            *configs.Config
	postgresDB      *gorm.DB
	orderRepository *repository.OrderRepository
	redisService    *RedisService
	walletClient    *client.WalletClient
}

func NewOrderService(
	config *configs.Config,
	postgresDB *gorm.DB,
	orderRepository *repository.OrderRepository,
	redisService *RedisService,
	walletClient *client.WalletClient) *OrderService {
	return &OrderService{
		conf:            config,
		postgresDB:      postgresDB,
		orderRepository: orderRepository,
		redisService:    redisService,
		walletClient:    walletClient,
	}
}

func (orderService *OrderService) Create(request request.OrderRequest) (response.OrderResponse, error) {
	valid, err := orderService.redisService.IdempotencyValidation(request.RequestId)
	if err != nil {
		return response.OrderResponse{}, err
	}

	if !valid {
		return response.OrderResponse{}, errors.New("idempotency validation error")
	}

	// todo
	//tx := orderService.getDbConnection()
	//orderEntity, err := orderService.orderRepository.Insert(tx, request)
	//if err != nil {
	//	return response.OrderResponse{}, err
	//}

	// fetch and validate product

	// call wallet-service
	//walletRequest := request2.WalletRequest{RequestID: request.RequestId, request.ProductId, request.}
	//orderService.walletClient.Charge()

	return response.OrderResponse{}, nil
}

func (orderService *OrderService) getDbConnection() *gorm.DB {
	tx := orderService.postgresDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	return tx
}
