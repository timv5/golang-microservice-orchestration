package repository

import "gorm.io/gorm"

type ProductRepositoryInterface interface {
	Exists(tx *gorm.DB, productId string) (bool, error)
}

type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

func (repo *ProductRepository) Exists(tx *gorm.DB, productId string) (bool, error) {
	var count int64
	err := tx.Table("products").Where("product_id = ?", productId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
