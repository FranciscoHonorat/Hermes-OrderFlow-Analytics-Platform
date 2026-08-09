package input

import (
	"context"

	"github.com/google/uuid"
)

type CreateOrderInput struct {
	CustomerID uuid.UUID              `json:"customer_id"`
	Items      []CreateOrderItemInput `json:"items"`
}

type CreateOrderItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"`
}

type OrderCommands interface {
	CreateOrder(ctx context.Context, input CreateOrderInput) (uuid.UUID, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error
	CancelOrder(ctx context.Context, orderID uuid.UUID) error
}
