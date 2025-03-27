package model

import (
	"time"
)

type Account struct {
	AccountId  string    `gorm:"type:bigint;primary_key" sql:"productOrderId"`
	Amount     float64   `gorm:"type:numeric;not null"`
	UpdateDate time.Time `gorm:"not null" sql:"createDate"`
}
