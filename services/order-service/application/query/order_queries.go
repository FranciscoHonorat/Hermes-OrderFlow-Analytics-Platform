package query

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OrderDTO struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     string
	Total      int64
	Currency   string
	Items      []ItemDTO
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ItemDTO struct {
	ProductID uuid.UUID
	Quantity  int
	Price     int64
}

type OrderQueries interface {
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderDTO, error)
	ListOrders(ctx context.Context, customerID *uuid.UUID, limit, offset int) ([]OrderDTO, error)
}
