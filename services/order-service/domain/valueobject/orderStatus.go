package valueobject

import (
	"encoding/json"
	"fmt"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusPlaced    OrderStatus = "PLACED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled, OrderStatusPaid, OrderStatusPlaced:
		return true
	default:
		return false
	}
}

func (s OrderStatus) String() string {
	return string(s)
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *OrderStatus) UnmarshalJSON(data []byte) error {
	var status string
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	orderStatus := OrderStatus(status)
	if !orderStatus.IsValid() {
		return fmt.Errorf("invalid order status: %w", domainErrors.ErrInvalidOrderStatus)
	}
	*s = orderStatus
	return nil
}
