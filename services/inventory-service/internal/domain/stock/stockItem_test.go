package stock_test

import (
	"testing"
	"time"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/domain-errors"
	stock "github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/stock"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedTime = time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

func newItem(t *testing.T, available, minimumQuantity int) *stock.StockItem {
	t.Helper()
	sku, err := valueobject.NewSKU("item1")
	require.NoError(t, err)

	item, err := stock.NewStockItem(sku, available, minimumQuantity, fixedTime)
	require.NoError(t, err)
	return item
}

func TestStockItem(t *testing.T) {
	sku, err := valueobject.NewSKU("item1")
	require.NoError(t, err)

	t.Run("Test NewStockItem method", func(t *testing.T) {
		tests := []struct {
			name            string
			sku             valueobject.SKU
			available       int
			minimumQuantity int
			expectedError   error
		}{
			{"Valid stock item", sku, 10, 2, nil},
			{"Invalid SKU", valueobject.SKU{}, 10, 2, domainErrors.ErrInvalidSKU},
			{"Negative available", sku, -1, 2, domainErrors.ErrInvalidQuantity},
			{"Negative minimum quantity", sku, 10, -1, domainErrors.ErrMinimumQuantity},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				item, err := stock.NewStockItem(tt.sku, tt.available, tt.minimumQuantity, fixedTime)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.sku, item.SKU())
					assert.Equal(t, tt.available, item.Available())
					assert.Equal(t, 0, item.Reserved())
					assert.Equal(t, tt.minimumQuantity, item.MinimumQuantity())
				}
			})
		}
	})

	t.Run("Test Reserve method", func(t *testing.T) {
		tests := []struct {
			name          string
			available     int
			minimum       int
			quantity      int
			expectedError error
		}{
			{"Valid quantity", 10, 2, 5, nil},
			{"Invalid quantity", 10, 2, -1, domainErrors.ErrInvalidQuantity},
			{"Insufficient stock", 10, 2, 15, domainErrors.ErrInsufficientStock},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				item := newItem(t, tt.available, tt.minimum)

				err := item.Reserve(tt.quantity, fixedTime)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.Equal(t, tt.available, item.Available(), "state must be unchanged on failure")
					assert.Equal(t, 0, item.Reserved(), "state must be unchanged on failure")
					assert.Empty(t, item.DomainEvents(), "no event must be emitted on failure")
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.available-tt.quantity, item.Available())
					assert.Equal(t, tt.quantity, item.Reserved())
				}
			})
		}

		t.Run("emits StockReserved event", func(t *testing.T) {
			item := newItem(t, 10, 2)

			require.NoError(t, item.Reserve(5, fixedTime))

			evts := item.PullEvents()
			require.Len(t, evts, 1)
			assert.Equal(t, "stock.reserved", evts[0].EventName())
			assert.Equal(t, "item1", evts[0].AggregateId())
			assert.Empty(t, item.DomainEvents(), "PullEvents must drain the event buffer")
		})

		t.Run("emits LowStockDetected when available drops below minimum", func(t *testing.T) {
			item := newItem(t, 10, 8)

			require.NoError(t, item.Reserve(5, fixedTime))

			evts := item.PullEvents()
			require.Len(t, evts, 2)
			assert.Equal(t, "stock.reserved", evts[0].EventName())
			assert.Equal(t, "stock.low_stock_detected", evts[1].EventName())
		})
	})

	t.Run("Test Release method", func(t *testing.T) {
		tests := []struct {
			name          string
			reserved      int
			quantity      int
			expectedError error
		}{
			{"Valid release", 5, 3, nil},
			{"Invalid quantity", 5, 0, domainErrors.ErrInvalidQuantity},
			{"Cannot release more than reserved", 3, 4, domainErrors.ErrNoReservedStock},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				item := newItem(t, 10, 2)
				require.NoError(t, item.Reserve(tt.reserved, fixedTime))
				item.ClearEvents()

				availableBeforeRelease := item.Available()

				err := item.Release(tt.quantity, fixedTime)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.Equal(t, availableBeforeRelease, item.Available(), "state must be unchanged on failure")
				} else {
					require.NoError(t, err)
					assert.Equal(t, availableBeforeRelease+tt.quantity, item.Available())
					assert.Equal(t, tt.reserved-tt.quantity, item.Reserved())

					evts := item.PullEvents()
					require.Len(t, evts, 1)
					assert.Equal(t, "stock.released", evts[0].EventName())
				}
			})
		}
	})

	t.Run("Test Replenish method", func(t *testing.T) {
		tests := []struct {
			name          string
			quantity      int
			expectedError error
		}{
			{"Valid replenishment", 5, nil},
			{"Invalid quantity", -5, domainErrors.ErrInvalidQuantity},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				item := newItem(t, 10, 2)

				err := item.Replenish(tt.quantity, fixedTime)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.Equal(t, 10, item.Available(), "state must be unchanged on failure")
				} else {
					require.NoError(t, err)
					assert.Equal(t, 10+tt.quantity, item.Available())

					evts := item.PullEvents()
					require.Len(t, evts, 1)
					assert.Equal(t, "stock.replenished", evts[0].EventName())
				}
			})
		}
	})
}
