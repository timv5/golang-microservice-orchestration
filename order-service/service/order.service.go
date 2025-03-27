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
	paymentClient     *client.PaymentClient
	productRepository *repository.ProductRepository
}

func NewOrderService(
	config *configs.Config,
	postgresDB *gorm.DB,
	orderRepository *repository.OrderRepository,
	redisService *RedisService,
	paymentClient *client.PaymentClient,
	productRepository *repository.ProductRepository) *OrderService {
	return &OrderService{
		conf:              config,
		postgresDB:        postgresDB,
		orderRepository:   orderRepository,
		redisService:      redisService,
		paymentClient:     paymentClient,
		productRepository: productRepository,
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

	// call payment-service
	uuid := uuid.NewV4().String()
	paymentRequest := client.PaymentRequest{
		RequestID: request.RequestId,
		UUID:      uuid, // new uuid is passed to payment service for the sake of idempotency
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
		// todo rollback on payment side
		tx.Rollback()
		return response.OrderResponse{}, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return response.OrderResponse{}, err
	}

	return response.OrderResponse{
		ProductOrderId: orderEntity.ProductOrderId,
		CreateDate:     orderEntity.CreateDate,
		ProductId:      orderEntity.ProductId,
		AccountId:      orderEntity.AccountId,
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
