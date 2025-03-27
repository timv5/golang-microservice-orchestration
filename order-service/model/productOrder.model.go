package model

import "time"

type ProductOrder struct {
	ProductOrderId string    `gorm:"type:bigint;primary_key" sql:"productOrderId"`
	ProductId      string    `gorm:"not null" sql:"productId"`
	AccountId      string    `gorm:"not null" sql:"accountId"`
	CreateDate     time.Time `gorm:"not null" sql:"createDate"`
}
