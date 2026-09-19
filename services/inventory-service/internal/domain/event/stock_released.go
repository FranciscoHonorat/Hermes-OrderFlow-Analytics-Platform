package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*StockReleased)(nil)

type StockReleased struct {
	events.BaseEvent
	Quantity       int
	AvailableAfter int
}

func NewStockReleased(sku string, quantity, availableAfter int, occurredAt time.Time) StockReleased {
	return StockReleased{
		BaseEvent:      events.NewBaseEvent("stock.released", sku, occurredAt),
		Quantity:       quantity,
		AvailableAfter: availableAfter,
	}
}
