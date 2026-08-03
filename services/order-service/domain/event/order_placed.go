package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
)

var _ DomainEvent = (*OrderPlaced)(nil)

type OrderPlaced struct {
	BaseEvent
	CustomerID  string
	TotalAmount valueobject.Money
	ItemCount   int
}

func NewOrderPlaced(orderID, customerID string, total valueobject.Money, item int, occurredAt time.Time) OrderPlaced {
	return OrderPlaced{
		BaseEvent:   NewBaseEvent("order.placed", orderID, occurredAt),
		CustomerID:  customerID,
		TotalAmount: total,
		ItemCount:   item,
	}
}
