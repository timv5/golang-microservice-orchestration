package repository

import (
	"gorm.io/gorm"
	"order-service/model"
)

type ProductRepositoryInterface interface {
	Fetch(tx *gorm.DB, productId string) (*model.Product, bool, error)
}

type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

func (repo *ProductRepository) Fetch(tx *gorm.DB, productId string) (*model.Product, bool, error) {
	var product model.Product
	err := tx.Table("products").Where("product_id = ?", productId).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &product, true, nil
}
