package handler

import (
	"net/http"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/input"
	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	stockQueries          input.StockQueries
	reserveStockUseCase   input.ReserveStockUseCase
	releaseStockUseCase   input.ReleaseStockUseCase
	replenishStockUseCase input.ReplenishStockUseCase
}

func NewStockHandler(
	queries input.StockQueries,
	reserve input.ReserveStockUseCase,
	release input.ReleaseStockUseCase,
	replenish input.ReplenishStockUseCase,
) *StockHandler {
	return &StockHandler{
		stockQueries:          queries,
		reserveStockUseCase:   reserve,
		releaseStockUseCase:   release,
		replenishStockUseCase: replenish,
	}
}

func (h *StockHandler) ReserveStock(c *gin.Context) {
	var req input.ReserveStockInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	if err := h.reserveStockUseCase.Execute(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock reserved successfully"})
}

func (h *StockHandler) ReleaseStock(c *gin.Context) {
	var req input.ReleaseStockInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	if err := h.releaseStockUseCase.Execute(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock released successfully"})
}

func (h *StockHandler) ReplenishStock(c *gin.Context) {
	var req input.ReplenishStockInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	if err := h.replenishStockUseCase.Execute(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock replenished successfully"})
}

func (h *StockHandler) GetStockBySKU(c *gin.Context) {
	sku := c.Param("sku")

	dto, err := h.stockQueries.GetStockBySKU(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock item not found"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

func (h *StockHandler) ListLowStock(c *gin.Context) {
	items, err := h.stockQueries.ListLowStock(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list low stock items"})
		return
	}

	c.JSON(http.StatusOK, items)
}
