package event_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderConfirmed(t *testing.T) {
	t.Run("should successfully create OrderConfirmed event", func(t *testing.T) {
		orderID, err := valueobject.NewOrderID(uuid.New())
		require.NoError(t, err)

		now := time.Now().UTC()
		orderConfirmed := event.NewOrderConfirmed(orderID.String(), now)

		require.Equal(t, "order.confirmed", orderConfirmed.EventName())
		require.Equal(t, orderID.String(), orderConfirmed.AggregateId())
		require.Equal(t, now, orderConfirmed.OccurredAt())
	})
}
