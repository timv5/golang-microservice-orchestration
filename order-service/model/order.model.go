package model

import "time"

type Order struct {
	ProductOrderId string    `gorm:"type:bigint;primary_key" sql:"productOrderId"`
	ProductId      string    `gorm:"not null" sql:"productId"`
	CreateDate     time.Time `gorm:"not null" sql:"createDate"`
}
