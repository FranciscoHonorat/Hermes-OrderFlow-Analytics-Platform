package event_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderShipped(t *testing.T) {
	t.Run("Test NewOrderShipped", func(t *testing.T) {
		orderID, err := valueobject.NewOrderID(uuid.New())
		require.NoError(t, err)

		shipmentID := uuid.New().String()
		carrier := "FedEx"
		trackingNumber := "123456789"
		now := time.Now().UTC()

		orderShipped := event.NewOrderShipped(
			orderID.String(),
			shipmentID,
			carrier,
			trackingNumber,
			now,
		)

		// Assert do BaseEvent
		require.Equal(t, "order.shipped", orderShipped.EventName())
		require.Equal(t, orderID.String(), orderShipped.AggregateId())
		require.Equal(t, now, orderShipped.OccurredAt())

		// Assert dos campos específicos de OrderShipped
		require.Equal(t, shipmentID, orderShipped.ShipmentID)
		require.Equal(t, carrier, orderShipped.Carrier)
		require.Equal(t, trackingNumber, orderShipped.TrackingNumber)
	})
}
