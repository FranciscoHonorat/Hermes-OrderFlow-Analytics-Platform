package input

import (
	"context"

	"github.com/google/uuid"
)

type CreateOrderItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"`
}

type CreateOrderInput struct {
	CustomerID uuid.UUID              `json:"customer_id"`
	Items      []CreateOrderItemInput `json:"items"`
}

type PlaceOrderUseCase interface {
	Execute(ctx context.Context, input CreateOrderInput) (uuid.UUID, error)
}

type ConfirmOrderUseCase interface {
	Execute(ctx context.Context, orderID uuid.UUID) error
}

type CancelOrderUseCase interface {
	Execute(ctx context.Context, orderID uuid.UUID) error
}

type ShipOrderUseCase interface {
	Execute(ctx context.Context, orderID uuid.UUID) error
}
