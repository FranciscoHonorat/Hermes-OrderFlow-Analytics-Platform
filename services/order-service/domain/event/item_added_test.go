package event_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestItemAdded(t *testing.T) {
	t.Run("Test NewItemAdded", func(t *testing.T) {
		orderID, err := valueobject.NewOrderID(uuid.New())
		require.NoError(t, err)

		productID, err := valueobject.NewProductID(uuid.New())
		require.NoError(t, err)

		quantity, err := valueobject.NewQuantity(2)
		require.NoError(t, err)

		unitPrice, err := valueobject.NewMoney(1000, "BRL")
		require.NoError(t, err)

		totalPrice, err := valueobject.NewMoney(2000, "BRL")
		require.NoError(t, err)

		tests := []struct {
			name        string
			orderID     valueobject.OrderID
			productID   valueobject.ProductID
			quantity    valueobject.Quantity
			unitPrice   valueobject.Money
			totalPrice  valueobject.Money
			expectError bool
		}{
			{
				name:        "Valid ItemAdded",
				orderID:     orderID,
				productID:   productID,
				quantity:    quantity,
				unitPrice:   unitPrice,
				totalPrice:  totalPrice,
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				itemAdded := event.NewOrderAdded(
					tt.orderID,
					tt.productID,
					tt.quantity,
					tt.unitPrice,
					tt.totalPrice,
					time.Now(),
				)

				err := itemAdded.ValidateOrderAdded()
				if tt.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}
