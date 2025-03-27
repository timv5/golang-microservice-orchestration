package repository

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"order-service/dto/request"
	"order-service/model"
	"time"
)

type OrderRepositoryInterface interface {
	Insert(tx *gorm.DB, order request.OrderRequest) (model.ProductOrder, error)
}

type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

func (repo *OrderRepository) Insert(tx *gorm.DB, order request.OrderRequest) (model.ProductOrder, error) {
	nowTime := time.Now()
	orderEntity := model.ProductOrder{
		ProductOrderId: uuid.NewV4().String(),
		ProductId:      order.ProductId,
		CreateDate:     nowTime,
		AccountId:      order.AccountID,
	}

	savedOrderEntity := tx.Create(&orderEntity)
	if savedOrderEntity.Error != nil {
		return model.ProductOrder{}, savedOrderEntity.Error
	}

	return orderEntity, nil
}
