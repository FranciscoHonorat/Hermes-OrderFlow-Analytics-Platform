package repository

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
)

type OrderRepository interface {
	Save(ctx context.Context, order *entity.Order) error
	FindByID(ctx context.Context, id valueobject.OrderID) (*entity.Order, error)
	GetOrdersByCustomerID(ctx context.Context, customerID valueobject.CustomerID) ([]*entity.Order, error)
}
