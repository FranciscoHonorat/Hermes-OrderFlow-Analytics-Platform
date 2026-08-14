package event

import "time"

var _ DomainEvent = (*OrderShipped)(nil)

type OrderShipped struct {
	BaseEvent
	ShipmentID     string
	Carrier        string
	TrackingNumber string
}

func NewOrderShipped(orderID, shipmentID, carrier, trackingNumber string, occurredAt time.Time) OrderShipped {
	return OrderShipped{
		BaseEvent:      NewBaseEvent("order.shipped", orderID, occurredAt),
		ShipmentID:     shipmentID,
		Carrier:        carrier,
		TrackingNumber: trackingNumber,
	}
}
