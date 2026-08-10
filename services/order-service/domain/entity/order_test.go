package entity_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	orderEntity "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder(t *testing.T) {
	t.Run("Test for NewOrder", func(t *testing.T) {
		tests := []struct {
			name        string
			id          uuid.UUID
			customerID  uuid.UUID
			expectedErr error
		}{
			{"Valid Order", uuid.New(), uuid.New(), nil},
			{"Invalid Order ID", uuid.Nil, uuid.New(), domainErrors.ErrInvalidOrderID},
			{"Invalid Customer ID", uuid.New(), uuid.Nil, domainErrors.ErrInvalidCustomerID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				orderID, err := valueobject.NewOrderID(tt.id)
				if tt.expectedErr == domainErrors.ErrInvalidOrderID {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				customerID, err := valueobject.NewCustomerID(tt.customerID)
				if tt.expectedErr == domainErrors.ErrInvalidCustomerID {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				_, err = orderEntity.NewOrder(orderID, customerID)
				assert.Equal(t, tt.expectedErr, err)
			})
		}
	})

	t.Run("Test for NewOrderMust", func(t *testing.T) {
		tests := []struct {
			name        string
			id          uuid.UUID
			customerID  uuid.UUID
			expectPanic bool
		}{
			{"Valid Order", uuid.New(), uuid.New(), false},
			{"Invalid Order ID", uuid.Nil, uuid.New(), true},
			{"Invalid Customer ID", uuid.New(), uuid.Nil, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				orderID, err := valueobject.NewOrderID(tt.id)
				if tt.expectPanic && tt.id == uuid.Nil {
					return
				}
				require.NoError(t, err)

				customerID, err := valueobject.NewCustomerID(tt.customerID)
				if tt.expectPanic && tt.customerID == uuid.Nil {
					return
				}
				require.NoError(t, err)

				if tt.expectPanic {
					assert.Panics(t, func() {
						orderEntity.NewOrderMust(orderID, customerID)
					})
				} else {
					assert.NotPanics(t, func() {
						orderEntity.NewOrderMust(orderID, customerID)
					})
				}
			})
		}
	})

	t.Run("Test for recalculateTotal", func(t *testing.T) {
		order := orderEntity.NewOrderMust(
			valueobject.NewOrderIDMust(uuid.New()),
			valueobject.NewCustomerIDMust(uuid.New()),
		)

		itemUSD := valueobject.NewOrderItemMust(
			valueobject.NewProductIDMust(uuid.New()),
			valueobject.NewMoneyMust(100, "USD"),
			valueobject.NewQuantityMust(1),
		)

		itemBRL := valueobject.NewOrderItemMust(
			valueobject.NewProductIDMust(uuid.New()),
			valueobject.NewMoneyMust(200, "BRL"),
			valueobject.NewQuantityMust(1),
		)

		require.NoError(t, order.AddItem(itemUSD))

		err := order.AddItem(itemBRL)
		if err != nil {
			require.Error(t, err)
		}

		itemBefore := len(order.Items())
		totalBefore := order.TotalPrice()

		err = order.AddItem(itemUSD)
		require.Error(t, err)

		itemAfter := len(order.Items())
		totalAfter := order.TotalPrice()

		assert.Equal(t, itemBefore, itemAfter, "Expected number of items to remain the same after failed addition")
		assert.Equal(t, totalBefore, totalAfter, "Expected total price to remain the same after failed addition")
	})

	t.Run("Test for AddItem", func(t *testing.T) {
		item := valueobject.NewOrderItemMust(
			valueobject.NewProductIDMust(uuid.New()),
			valueobject.NewMoneyMust(100, "USD"),
			valueobject.NewQuantityMust(1),
		)

		order := orderEntity.NewOrderMust(
			valueobject.NewOrderIDMust(uuid.New()),
			valueobject.NewCustomerIDMust(uuid.New()),
		)

		assert.NotPanics(t, func() {
			err := order.AddItem(item)
			require.NoError(t, err)
		})
	})

	t.Run("Test for UpdateStatus", func(t *testing.T) {
		tests := []struct {
			name        string
			newStatus   valueobject.OrderStatus
			expectedErr error
		}{
			{"Valid Status Update", valueobject.OrderStatusPaid, nil},
			{"Invalid Status Update", "invalid_status", domainErrors.ErrInvalidOrderStatus},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				order := orderEntity.NewOrderMust(
					valueobject.NewOrderIDMust(uuid.New()),
					valueobject.NewCustomerIDMust(uuid.New()),
				)

				err := order.UpdateStatus(tt.newStatus)
				assert.Equal(t, tt.expectedErr, err)
			})
		}
	})
}
