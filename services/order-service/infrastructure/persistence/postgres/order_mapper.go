package postgres

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
)

type OrderRow struct {
	ID         string    `db:"id"`
	CustomerID string    `db:"customer_id"`
	TotalCents int64     `db:"total_cents"`
	Status     string    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type OrderMapper struct{}

func NewOrderMapper() *OrderMapper {
	return &OrderMapper{}
}

func (m *OrderMapper) ToPersistence(order *entity.Order) *OrderRow {
	if order == nil {
		return nil
	}

	return &OrderRow{
		ID:         order.OrderID().String(),
		CustomerID: order.CustomerID().String(),
		TotalCents: order.TotalPrice().Amount(),
		Status:     order.Status().String(),
		CreatedAt:  order.CreatedAt(),
		UpdatedAt:  order.UpdatedAt(),
	}
}

func (m *OrderMapper) ToDomain(row *OrderRow) (*entity.Order, error) {
	if row == nil {
		return nil, nil
	}

	orderID, err := valueobject.ParseOrderID(row.ID)
	if err != nil {
		return nil, err
	}

	customerID, err := valueobject.ParseCustomerID(row.CustomerID)
	if err != nil {
		return nil, err
	}

	money, err := valueobject.NewMoneyFromAmount(row.TotalCents)
	if err != nil {
		return nil, err
	}

	order, err := entity.RestoreOrder(
		orderID,
		customerID,
		money,
		[]valueobject.OrderItem{},
		valueobject.Address{},
		valueobject.OrderStatus(row.Status),
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return order, nil
}
