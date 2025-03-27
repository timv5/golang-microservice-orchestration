package repository

import (
	"gorm.io/gorm"
	"payment-service/model"
)

type TransactionRepositoryInterface interface {
	Insert(tx *gorm.DB, transaction *model.Transaction) error
}

type TransactionRepository struct{}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{}
}

func (r *TransactionRepository) Insert(tx *gorm.DB, transaction *model.Transaction) error {
	return tx.Create(transaction).Error
}
