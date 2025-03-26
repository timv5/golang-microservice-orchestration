package repository

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"order-service/dto/request"
	"order-service/model"
	"time"
)

type OrderRepositoryInterface interface {
	Insert(tx *gorm.DB, order request.OrderRequest) (model.Order, error)
}

type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

func (repo *OrderRepository) Insert(tx *gorm.DB, order request.OrderRequest) (model.Order, error) {
	nowTime := time.Now()
	orderEntity := model.Order{
		ProductOrderId: uuid.NewV4().String(),
		ProductId:      order.ProductId,
		CreateDate:     nowTime,
	}

	savedOrderEntity := tx.Create(&orderEntity)
	if savedOrderEntity.Error != nil {
		return model.Order{}, savedOrderEntity.Error
	}

	return orderEntity, nil
}
