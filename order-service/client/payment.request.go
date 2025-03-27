package client

type PaymentRequest struct {
	RequestID string  `json:"requestId"`
	UUID      string  `json:"uuid"`
	ProductId string  `json:"productId" binding:"required"`
	Amount    float64 `json:"amount"`
	AccountID string  `json:"accountId"`
}
