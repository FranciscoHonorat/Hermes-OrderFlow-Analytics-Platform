package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*OrderConfirmed)(nil)

type OrderConfirmed struct {
	events.BaseEvent
}

func NewOrderConfirmed(orderID string, occurredAt time.Time) OrderConfirmed {
	return OrderConfirmed{
		BaseEvent: events.NewBaseEvent("order.confirmed", orderID, occurredAt),
	}
}
