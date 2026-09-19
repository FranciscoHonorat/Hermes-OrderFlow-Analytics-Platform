package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*OrderCancelled)(nil)

type OrderCancelled struct {
	events.BaseEvent
	CustomerID string
	Reason     string
}

func NewOrderCancelled(orderID, customerID, reason string, occurredAt time.Time) OrderCancelled {
	return OrderCancelled{
		BaseEvent:  events.NewBaseEvent("order.cancelled", orderID, occurredAt),
		CustomerID: customerID,
		Reason:     reason,
	}
}
