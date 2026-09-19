package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*LowStockDetected)(nil)

type LowStockDetected struct {
	events.BaseEvent
	Available       int
	MinimumQuantity int
}

func NewLowStockDetected(sku string, available, minimumQuantity int, occurredAt time.Time) LowStockDetected {
	return LowStockDetected{
		BaseEvent:       events.NewBaseEvent("stock.low_stock_detected", sku, occurredAt),
		Available:       available,
		MinimumQuantity: minimumQuantity,
	}
}
