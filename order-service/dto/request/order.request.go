package request

type OrderRequest struct {
	ProductId string `json:"productId"  binding:"required"`
	RequestId string `json:"requestId" binding:"required"`
}
