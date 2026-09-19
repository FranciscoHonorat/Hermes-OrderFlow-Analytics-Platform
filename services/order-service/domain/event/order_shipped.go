package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*OrderShipped)(nil)

type OrderShipped struct {
	events.BaseEvent
	ShipmentID     string
	Carrier        string
	TrackingNumber string
}

func NewOrderShipped(orderID, shipmentID, carrier, trackingNumber string, occurredAt time.Time) OrderShipped {
	return OrderShipped{
		BaseEvent:      events.NewBaseEvent("order.shipped", orderID, occurredAt),
		ShipmentID:     shipmentID,
		Carrier:        carrier,
		TrackingNumber: trackingNumber,
	}
}
