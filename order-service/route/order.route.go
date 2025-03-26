package route

import (
	"github.com/gin-gonic/gin"
	"order-service/handler"
)

type OrderRouteHandler struct {
	orderHandler handler.OrderHandler
}

func NewOrderRouteHandler(orderHandler handler.OrderHandler) OrderRouteHandler {
	return OrderRouteHandler{
		orderHandler: orderHandler,
	}
}

func (h *OrderRouteHandler) OrderRoute(group *gin.RouterGroup) {
	router := group.Group("order")
	router.POST("/create", h.orderHandler.MakeOrder)
}
