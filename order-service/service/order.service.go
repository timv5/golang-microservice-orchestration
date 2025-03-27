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
	conf              *configs.Config
	postgresDB        *gorm.DB
	orderRepository   *repository.OrderRepository
	redisService      *RedisService
	walletClient      *client.WalletClient
	productRepository *repository.ProductRepository
}

func NewOrderService(
	config *configs.Config,
	postgresDB *gorm.DB,
	orderRepository *repository.OrderRepository,
	redisService *RedisService,
	walletClient *client.WalletClient,
	productRepository *repository.ProductRepository) *OrderService {
	return &OrderService{
		conf:              config,
		postgresDB:        postgresDB,
		orderRepository:   orderRepository,
		redisService:      redisService,
		walletClient:      walletClient,
		productRepository: productRepository,
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

	tx := orderService.getDbConnection()
	exists, err := orderService.productRepository.Exists(tx, request.ProductId)
	if err != nil {
		return response.OrderResponse{}, err
	}

	if !exists {
		return response.OrderResponse{}, errors.New("product does not exist")
	}

	orderEntity, err := orderService.orderRepository.Insert(tx, request)
	if err != nil {
		return response.OrderResponse{}, err
	}

	// call wallet-service
	//walletRequest := request2.WalletRequest{RequestID: request.RequestId, request.ProductId, request.}
	//orderService.walletClient.Charge()

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return response.OrderResponse{}, err
	}

	return response.OrderResponse{
		ProductOrderId: orderEntity.ProductOrderId,
		CreateDate:     orderEntity.CreateDate,
		ProductId:      orderEntity.ProductId,
	}, nil
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
