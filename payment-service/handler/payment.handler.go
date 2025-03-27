package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"payment-service/configs"
	"payment-service/dto/request"
	"payment-service/service"
)

type PaymentHandler struct {
	postgresDB     *gorm.DB
	paymentService *service.PaymentService
	config         *configs.Config
}

type PaymentHandlerInterface interface {
	MakeOrder(ctx *gin.Context)
}

func NewPaymentHandler(postgresDB *gorm.DB, paymentService *service.PaymentService, config *configs.Config) PaymentHandler {
	return PaymentHandler{
		postgresDB:     postgresDB,
		paymentService: paymentService,
		config:         config,
	}
}

func (paymentHandler PaymentHandler) ProcessPayment(ctx *gin.Context) {
	var paymentRequest request.PaymentRequest
	if err := ctx.ShouldBindJSON(&paymentRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	if paymentRequest.RequestID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "wrong orderRequest params, missing request id"})
		return
	}

	if paymentRequest.ProductId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "wrong orderRequest params, missing product"})
		return
	}

	if paymentRequest.AccountID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "wrong orderRequest params, missing account"})
		return
	}

	err := paymentHandler.paymentService.ProcessPayment(paymentRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
