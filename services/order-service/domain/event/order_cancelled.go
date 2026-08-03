package event

import "time"

var _ DomainEvent = (*OrderPlaced)(nil)

type OrderCancelled struct {
	BaseEvent
	CustomerID string
	Reason     string
}

func NewOrderCancelled(orderID, customerID, reason string, occurredAt time.Time) OrderCancelled {
	return OrderCancelled{
		BaseEvent:  NewBaseEvent("order.cancelled", orderID, occurredAt),
		CustomerID: customerID,
		Reason:     reason,
	}
}
