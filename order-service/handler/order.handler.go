package handler

import (
	"net/http"
	"order-service/configs"
	"order-service/dto/request"
	"order-service/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	postgresDB   *gorm.DB
	orderService *service.OrderService
	config       *configs.Config
}

type OrderHandlerInterface interface {
	MakeOrder(ctx *gin.Context)
}

func NewOrderHandler(postgresDB *gorm.DB, orderService *service.OrderService, config *configs.Config) OrderHandler {
	return OrderHandler{
		postgresDB:   postgresDB,
		orderService: orderService,
		config:       config,
	}
}

func (orderHandler OrderHandler) MakeOrder(ctx *gin.Context) {
	var orderRequest request.OrderRequest
	if err := ctx.ShouldBindJSON(&orderRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	if orderRequest.RequestId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "wrong orderRequest params"})
		return
	}

	response, err := orderHandler.orderService.Create(orderRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
