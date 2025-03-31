package model

import (
	"time"
)

type Transaction struct {
	TransactionId string    `gorm:"type:bigint;primary_key" sql:"productOrderId"`
	ProductId     string    `gorm:"not null" sql:"productId"`
	Amount        float64   `gorm:"type:numeric;not null"`
	CreateDate    time.Time `gorm:"not null" sql:"createDate"`
	RequestId     string    `gorm:"not null" sql:"requestId"`
	AccountId     string    `gorm:"not null" sql:"accountId"`
}
