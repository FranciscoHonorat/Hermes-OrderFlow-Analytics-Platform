package event_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderCancelled(t *testing.T) {
	t.Run("Test NewOrderCancelled", func(t *testing.T) {
		orderID, err := valueobject.NewOrderID(uuid.New())
		require.NoError(t, err)

		customerID, err := valueobject.NewCustomerID(uuid.New())
		require.NoError(t, err)

		reason := "Customer requested cancellation"

		orderCancelled := event.NewOrderCancelled(orderID.String(), customerID.String(), reason, time.Now())

		require.Equal(t, "order.cancelled", orderCancelled.EventName())
		require.Equal(t, orderID.String(), orderCancelled.AggregateId())
		require.Equal(t, customerID.String(), orderCancelled.CustomerID)
		require.Equal(t, reason, orderCancelled.Reason)
	})
}
