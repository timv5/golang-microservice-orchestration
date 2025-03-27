package model

import "time"

type Product struct {
	ProductId  string    `gorm:"type:bigint;primary_key" sql:"productOrderId"`
	Name       string    `gorm:"not null" sql:"productId"`
	Price      float64   `gorm:"type:numeric;not null" sql:"accountId"`
	CreateDate time.Time `gorm:"not null" sql:"createDate"`
	UpdateDate time.Time `gorm:"not null" sql:"createDate"`
}
