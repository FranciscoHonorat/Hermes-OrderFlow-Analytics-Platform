package input

import "context"

type ReserveStockInput struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

type ReserveStockUseCase interface {
	Execute(ctx context.Context, in ReserveStockInput) error
}

type ReleaseStockInput struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

type ReleaseStockUseCase interface {
	Execute(ctx context.Context, in ReleaseStockInput) error
}

type ReplenishStockInput struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

type ReplenishStockUseCase interface {
	Execute(ctx context.Context, in ReplenishStockInput) error
}
