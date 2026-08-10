package valueobject_test

import (
	"testing"

	order "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderItem(t *testing.T) {
	t.Run("Test for NewOrderItem", func(t *testing.T) {
		tests := []struct {
			name          string
			productID     order.ProductID
			unitPrice     order.Money
			quantity      order.Quantity
			expectedError error
		}{
			{
				name:          "Valid OrderItem",
				productID:     order.NewProductIDMust(uuid.New()),
				unitPrice:     order.NewMoneyMust(100, "USD"),
				quantity:      order.NewQuantityMust(2),
				expectedError: nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p, err := order.NewOrderItem(tt.productID, tt.unitPrice, tt.quantity)

				assert.ErrorIs(t, err, tt.expectedError)

				if tt.expectedError == nil {
					require.NotNil(t, p)
					assert.Equal(t, tt.productID, p.ProductID())
					assert.True(t, p.UnitPrice().Equals(tt.unitPrice))
					assert.Equal(t, tt.quantity, p.Quantity())
				}
			})
		}
	})

	t.Run("Test for subTotal", func(t *testing.T) {
		tests := []struct {
			name          string
			productID     order.ProductID
			unitPrice     order.Money
			quantity      order.Quantity
			expectedTotal order.Money
		}{
			{
				name:          "Valid OrderItem",
				productID:     order.NewProductIDMust(uuid.New()),
				unitPrice:     order.NewMoneyMust(100, "USD"),
				quantity:      order.NewQuantityMust(2),
				expectedTotal: order.NewMoneyMust(200, "USD"),
			},
			{
				name:          "Valid OrderItem with different currency",
				productID:     order.NewProductIDMust(uuid.New()),
				unitPrice:     order.NewMoneyMust(50, "BRL"),
				quantity:      order.NewQuantityMust(3),
				expectedTotal: order.NewMoneyMust(150, "BRL"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p, err := order.NewOrderItem(tt.productID, tt.unitPrice, tt.quantity)
				require.NoError(t, err)

				total, err := p.SubTotal()
				require.NoError(t, err)

				assert.True(t, total.Equals(tt.expectedTotal), "Expected SubTotal %v, got %v", tt.expectedTotal, total)
			})
		}
	})

	t.Run("Test for MarshalJSON", func(t *testing.T) {
		pid := order.NewProductIDMust(uuid.New())
		tests := []struct {
			name         string
			productID    order.ProductID
			unitPrice    order.Money
			quantity     order.Quantity
			expectedJSON string
		}{
			{
				name:         "Valid OrderItem",
				productID:    pid,
				unitPrice:    order.NewMoneyMust(100, "USD"),
				quantity:     order.NewQuantityMust(2),
				expectedJSON: `{"product_id":"` + pid.String() + `","unit_price":{"amount":100,"currency":"USD"},"quantity":2}`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p, err := order.NewOrderItem(tt.productID, tt.unitPrice, tt.quantity)
				require.NoError(t, err)

				jsonData, err := p.MarshalJSON()
				require.NoError(t, err)

				assert.JSONEq(t, tt.expectedJSON, string(jsonData))
			})
		}
	})

	t.Run("Test for UnmarshalJSON", func(t *testing.T) {
		pid := order.NewProductIDMust(uuid.New())
		tests := []struct {
			name         string
			jsonData     string
			expectedItem order.OrderItem
			expectError  bool
		}{
			{
				name:         "Valid OrderItem",
				jsonData:     `{"product_id":"` + pid.String() + `","unit_price":{"amount":100,"currency":"USD"},"quantity":2}`,
				expectedItem: order.NewOrderItemMust(pid, order.NewMoneyMust(100, "USD"), order.NewQuantityMust(2)),
				expectError:  false,
			},
			{
				name:         "Invalid JSON Syntax",
				jsonData:     `{"product_id":"invalid-uuid",}`,
				expectedItem: order.OrderItem{},
				expectError:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var item order.OrderItem
				err := item.UnmarshalJSON([]byte(tt.jsonData))

				if tt.expectError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedItem.ProductID(), item.ProductID())
					assert.True(t, item.UnitPrice().Equals(tt.expectedItem.UnitPrice()))
					assert.Equal(t, tt.expectedItem.Quantity(), item.Quantity())
				}
			})
		}
	})
}
