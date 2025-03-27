package route

import (
	"github.com/gin-gonic/gin"
	"payment-service/handler"
)

type PaymentRouteHandler struct {
	paymentHandler handler.PaymentHandler
}

func NewPaymentRouteHandler(paymentHandler handler.PaymentHandler) PaymentRouteHandler {
	return PaymentRouteHandler{
		paymentHandler: paymentHandler,
	}
}

func (h *PaymentRouteHandler) PaymentRoute(group *gin.RouterGroup) {
	router := group.Group("payment")
	router.POST("/process", h.paymentHandler.ProcessPayment)
}
