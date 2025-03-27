package request

type PaymentRequest struct {
	RequestID string  `json:"request_id"`
	ProductId string  `json:"product_id" binding:"required"`
	Amount    float64 `json:"amount"`
	AccountID string  `json:"account_id"`
}
