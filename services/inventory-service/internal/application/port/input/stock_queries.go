package input

import (
	"context"
	"time"
)

type StockDTO struct {
	SKU             string    `json:"sku"`
	Available       int       `json:"available"`
	Reserved        int       `json:"reserved"`
	MinimumQuantity int       `json:"minimum_quantity"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StockQueries interface {
	GetStockBySKU(ctx context.Context, sku string) (*StockDTO, error)
	ListLowStock(ctx context.Context) ([]StockDTO, error)
}
