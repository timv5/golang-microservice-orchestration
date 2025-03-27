package repository

import (
	"errors"
	"gorm.io/gorm"
	"payment-service/model"
)

type AccountRepositoryInterface interface {
	DeductAmount(tx *gorm.DB, accountId string, amount float64) error
}

type AccountRepository struct{}

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{}
}

func (r *AccountRepository) DeductAmount(tx *gorm.DB, accountId string, amount float64) error {
	var account model.Account
	if err := tx.Where("account_id = ?", accountId).First(&account).Error; err != nil {
		return err
	}

	if account.Amount < amount {
		return errors.New("insufficient balance")
	}

	account.Amount -= amount
	account.UpdateDate = account.UpdateDate.UTC()

	if err := tx.Save(&account).Error; err != nil {
		return err
	}

	return nil
}
