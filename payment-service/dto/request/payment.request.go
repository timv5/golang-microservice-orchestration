package request

type PaymentRequest struct {
	RequestID string  `json:"requestId"`
	ProductId string  `json:"productId" binding:"required"`
	Amount    float64 `json:"amount"`
	AccountID string  `json:"accountId"`
}
