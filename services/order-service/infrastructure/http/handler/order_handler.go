package handler

import (
	"net/http"
	"strconv"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/input"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderIDRequest struct {
	OrderID string `json:"order_id" binding:"required"`
}

type OrderHandler struct {
	orderQueries        input.OrderQueries
	placeOrderUseCase   input.PlaceOrderUseCase
	confirmOrderUseCase input.ConfirmOrderUseCase
	cancelOrderUseCase  input.CancelOrderUseCase
	shipOrderUseCase    input.ShipOrderUseCase
}

func NewOrderHandler(
	queries input.OrderQueries,
	place input.PlaceOrderUseCase,
	confirm input.ConfirmOrderUseCase,
	cancel input.CancelOrderUseCase,
	ship input.ShipOrderUseCase,
) *OrderHandler {
	return &OrderHandler{
		orderQueries:        queries,
		placeOrderUseCase:   place,
		confirmOrderUseCase: confirm,
		cancelOrderUseCase:  cancel,
		shipOrderUseCase:    ship,
	}
}

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req input.CreateOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	resultID, err := h.placeOrderUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order_id": resultID.String()})
}

func (h *OrderHandler) ConfirmOrder(c *gin.Context) {
	var req OrderIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	if err := h.confirmOrderUseCase.Execute(c.Request.Context(), orderUUID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order confirmed successfully"})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	var req OrderIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	if err := h.cancelOrderUseCase.Execute(c.Request.Context(), orderUUID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

func (h *OrderHandler) ShipOrder(c *gin.Context) {
	var req OrderIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	if err := h.shipOrderUseCase.Execute(c.Request.Context(), orderUUID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order shipped successfully"})
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id") // Mudado de Query Param para Path Param (padrão RESTful elegante)
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	orderDTO, err := h.orderQueries.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, orderDTO)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var customerID *uuid.UUID
	if cidStr := c.Query("customer_id"); cidStr != "" {
		if cid, err := uuid.Parse(cidStr); err == nil {
			customerID = &cid
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID format"})
			return
		}
	}

	orders, err := h.orderQueries.ListOrders(c.Request.Context(), customerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
