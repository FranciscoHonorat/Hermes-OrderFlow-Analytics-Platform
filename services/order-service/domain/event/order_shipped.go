package event

import "time"

var _ DomainEvent = (*OrderShipped)(nil)

type OrderShipped struct {
	BaseEvent
}

func NewOrderShipped(orderID string, occurredAt time.Time) OrderShipped {
	return OrderShipped{
		BaseEvent: NewBaseEvent("order.shipped", orderID, occurredAt),
	}
}
