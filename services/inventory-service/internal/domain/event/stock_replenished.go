package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*StockReplenished)(nil)

type StockReplenished struct {
	events.BaseEvent
	Quantity       int
	AvailableAfter int
}

func NewStockReplenished(sku string, quantity, availableAfter int, occurredAt time.Time) StockReplenished {
	return StockReplenished{
		BaseEvent:      events.NewBaseEvent("stock.replenished", sku, occurredAt),
		Quantity:       quantity,
		AvailableAfter: availableAfter,
	}
}
