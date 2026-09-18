package stock_test

import (
	"testing"

	stock "github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/stock"
	"github.com/stretchr/testify/assert"
)

func TestStockItem(t *testing.T) {
	t.Run("Functionality", func(t *testing.T) {
		t.Run("Reserve", func(t *testing.T) {
			t.Run("valid quantity", func(t *testing.T) {
				item := &stock.StockItem{
					SKU:             "item1",
					Available:       10,
					Reserved:        0,
					MinimumQuantity: 2,
				}

				err := item.Reserve(5)
				assert.NoError(t, err)
				assert.Equal(t, 5, item.Available)
				assert.Equal(t, 5, item.Reserved)
			})

			t.Run("invalid quantity", func(t *testing.T) {
				item := &stock.StockItem{
					SKU:             "item1",
					Available:       10,
					Reserved:        0,
					MinimumQuantity: 2,
				}
				err := item.Reserve(-1)
				assert.Equal(t, stock.ErrInvalidQuantity, err)
			})

			t.Run("insufficient stock", func(t *testing.T) {
				item := &stock.StockItem{
					SKU:             "item1",
					Available:       10,
					Reserved:        0,
					MinimumQuantity: 2,
				}
				err := item.Reserve(15)
				assert.Equal(t, stock.ErrInsufficientStock, err)
			})

			t.Run("failed reservation keeps state unchanged", func(t *testing.T) {
				item := &stock.StockItem{
					SKU:             "item1",
					Available:       10,
					Reserved:        0,
					MinimumQuantity: 2,
				}
				err := item.Reserve(15)
				assert.Equal(t, stock.ErrInsufficientStock, err)
				assert.Equal(t, 10, item.Available)
				assert.Equal(t, 0, item.Reserved)
			})
		})
	})
}
