package service

import (
	"errors"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"net/http"
	"order-service/client"
	"order-service/configs"
	"order-service/dto/request"
	"order-service/dto/response"
	"order-service/orchestration"
	"order-service/repository"
)

type OrderServiceInterface interface {
	Create(request request.OrderRequest) (response.OrderResponse, error)
}

type OrderService struct {
	conf                 *configs.Config
	postgresDB           *gorm.DB
	orderRepository      *repository.OrderRepository
	redisService         *RedisService
	paymentClient        *client.PaymentClient
	productRepository    *repository.ProductRepository
	orchestrationManager *orchestration.Manager
}

func NewOrderService(
	config *configs.Config,
	postgresDB *gorm.DB,
	orderRepository *repository.OrderRepository,
	redisService *RedisService,
	paymentClient *client.PaymentClient,
	productRepository *repository.ProductRepository,
	orchestrationManager *orchestration.Manager) *OrderService {
	return &OrderService{
		conf:                 config,
		postgresDB:           postgresDB,
		orderRepository:      orderRepository,
		redisService:         redisService,
		paymentClient:        paymentClient,
		productRepository:    productRepository,
		orchestrationManager: orchestrationManager,
	}
}

func (os *OrderService) Create(request request.OrderRequest) (response.OrderResponse, error) {
	valid, err := os.redisService.IdempotencyValidation(request.RequestId)
	if err != nil {
		return response.OrderResponse{}, err
	}
	if !valid {
		return response.OrderResponse{}, errors.New("idempotency validation error")
	}

	err = os.orchestrationManager.Start(request.RequestId)
	if err != nil {
		return response.OrderResponse{}, err
	}

	tx := os.getDbConnection()
	product, exists, err := os.productRepository.Fetch(tx, request.ProductId)
	if err != nil {
		tx.Rollback()
		return response.OrderResponse{}, err
	}

	if !exists {
		tx.Rollback()
		return response.OrderResponse{}, errors.New("product does not exist")
	}

	orderEntity, err := os.orderRepository.Insert(tx, request)
	if err != nil {
		tx.Rollback()
		return response.OrderResponse{}, err
	}

	paymentRequest := client.PaymentRequest{
		RequestID: request.RequestId,
		UUID:      uuid.NewV4().String(),
		ProductId: request.ProductId,
		AccountID: request.AccountID,
		Amount:    product.Price,
	}
	statusCode, err := os.paymentClient.Process(paymentRequest)
	if err != nil {
		tx.Rollback()
		return response.OrderResponse{}, err
	}

	if statusCode != http.StatusOK {
		tx.Rollback()
		return response.OrderResponse{}, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		err := os.orchestrationManager.Rollback(request.RequestId)
		if err != nil {
			return response.OrderResponse{}, err
		}
		return response.OrderResponse{}, err
	}

	err = os.orchestrationManager.End(request.RequestId)
	if err != nil {
		return response.OrderResponse{}, err
	}

	return response.OrderResponse{
		ProductOrderId: orderEntity.ProductOrderId,
		CreateDate:     orderEntity.CreateDate,
		ProductId:      orderEntity.ProductId,
		AccountId:      orderEntity.AccountId,
	}, nil
}

func (os *OrderService) getDbConnection() *gorm.DB {
	tx := os.postgresDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	return tx
}
