package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*StockReserved)(nil)

type StockReserved struct {
	events.BaseEvent
	Quantity       int
	AvailableAfter int
}

func NewStockReserved(sku string, quantity, availableAfter int, occurredAt time.Time) StockReserved {
	return StockReserved{
		BaseEvent:      events.NewBaseEvent("stock.reserved", sku, occurredAt),
		Quantity:       quantity,
		AvailableAfter: availableAfter,
	}
}
