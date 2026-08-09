package input

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OrderDTO struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Status     string    `json:"status"`
	Total      float64   `json:"total"`
	Currency   string    `json:"currency"`
	Items      []ItemDTO `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ItemDTO struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"`
}

type OrderQueries interface {
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderDTO, error)
	ListOrders(ctx context.Context, customerID *uuid.UUID, limit, offset int) ([]OrderDTO, error)
}
