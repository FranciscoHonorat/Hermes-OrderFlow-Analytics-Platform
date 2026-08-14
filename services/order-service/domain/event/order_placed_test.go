package event_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderPlaced(t *testing.T) {
	t.Run("Test NewOrderPlaced", func(t *testing.T) {
		orderID, err := valueobject.NewOrderID(uuid.New())
		require.NoError(t, err)

		customerID, err := valueobject.NewCustomerID(uuid.New())
		require.NoError(t, err)

		totalAmount, err := valueobject.NewMoney(1000, "BRL")
		require.NoError(t, err)

		itemCount := 3

		orderPlaced := event.NewOrderPlaced(orderID.String(), customerID.String(), totalAmount, itemCount, time.Now())

		require.Equal(t, "order.placed", orderPlaced.EventName())
		require.Equal(t, orderID.String(), orderPlaced.AggregateId())
		require.Equal(t, customerID.String(), orderPlaced.CustomerID)
		require.Equal(t, totalAmount, orderPlaced.TotalAmount)
		require.Equal(t, itemCount, orderPlaced.ItemCount)
	})
}
