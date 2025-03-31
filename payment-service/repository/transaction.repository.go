package repository

import (
	"gorm.io/gorm"
	"payment-service/model"
)

type TransactionRepositoryInterface interface {
	Insert(tx *gorm.DB, transaction *model.Transaction) error
	Delete(tx *gorm.DB, requestId string) (*model.Transaction, error)
}

type TransactionRepository struct{}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{}
}

func (r *TransactionRepository) Insert(tx *gorm.DB, transaction *model.Transaction) error {
	return tx.Create(transaction).Error
}

func (r *TransactionRepository) Delete(tx *gorm.DB, requestId string) (*model.Transaction, error) {
	var transaction model.Transaction
	if err := tx.Where("request_id = ?", requestId).First(&transaction).Error; err != nil {
		return nil, err
	}
	if err := tx.Delete(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}
