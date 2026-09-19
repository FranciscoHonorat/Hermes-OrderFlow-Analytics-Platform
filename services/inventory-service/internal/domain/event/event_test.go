package event_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/event"
	"github.com/stretchr/testify/require"
)

func TestStockEvents(t *testing.T) {
	now := time.Now().UTC()

	t.Run("Test NewStockReserved", func(t *testing.T) {
		evt := event.NewStockReserved("sku-1", 5, 10, now)

		require.Equal(t, "stock.reserved", evt.EventName())
		require.Equal(t, "sku-1", evt.AggregateId())
		require.True(t, evt.OccurredAt().Equal(now))
		require.Equal(t, 5, evt.Quantity)
		require.Equal(t, 10, evt.AvailableAfter)
	})

	t.Run("Test NewStockReleased", func(t *testing.T) {
		evt := event.NewStockReleased("sku-1", 3, 13, now)

		require.Equal(t, "stock.released", evt.EventName())
		require.Equal(t, "sku-1", evt.AggregateId())
		require.True(t, evt.OccurredAt().Equal(now))
		require.Equal(t, 3, evt.Quantity)
		require.Equal(t, 13, evt.AvailableAfter)
	})

	t.Run("Test NewStockReplenished", func(t *testing.T) {
		evt := event.NewStockReplenished("sku-1", 20, 30, now)

		require.Equal(t, "stock.replenished", evt.EventName())
		require.Equal(t, "sku-1", evt.AggregateId())
		require.True(t, evt.OccurredAt().Equal(now))
		require.Equal(t, 20, evt.Quantity)
		require.Equal(t, 30, evt.AvailableAfter)
	})

	t.Run("Test NewLowStockDetected", func(t *testing.T) {
		evt := event.NewLowStockDetected("sku-1", 1, 5, now)

		require.Equal(t, "stock.low_stock_detected", evt.EventName())
		require.Equal(t, "sku-1", evt.AggregateId())
		require.True(t, evt.OccurredAt().Equal(now))
		require.Equal(t, 1, evt.Available)
		require.Equal(t, 5, evt.MinimumQuantity)
	})
}
