package response

import "time"

type OrderResponse struct {
	ProductOrderId string
	CreateDate     time.Time
	ProductId      string
	AccountId      string
}
