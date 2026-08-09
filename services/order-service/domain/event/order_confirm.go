package event

import "time"

var _ DomainEvent = (*OrderConfirmed)(nil)

type OrderConfirmed struct {
	BaseEvent
}

func NewOrderConfirmed(orderID string, occurredAt time.Time) OrderConfirmed {
	return OrderConfirmed{
		BaseEvent: NewBaseEvent("order.confirmed", orderID, occurredAt),
	}
}
