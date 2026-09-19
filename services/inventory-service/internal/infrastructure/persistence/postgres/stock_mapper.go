package postgres

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/stock"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
)

type StockItemRow struct {
	SKU             string    `db:"sku"`
	Available       int       `db:"available"`
	Reserved        int       `db:"reserved"`
	MinimumQuantity int       `db:"minimum_quantity"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type StockMapper struct{}

func NewStockMapper() *StockMapper {
	return &StockMapper{}
}

func (m *StockMapper) ToPersistence(item *stock.StockItem) *StockItemRow {
	if item == nil {
		return nil
	}

	return &StockItemRow{
		SKU:             item.SKU().String(),
		Available:       item.Available(),
		Reserved:        item.Reserved(),
		MinimumQuantity: item.MinimumQuantity(),
		UpdatedAt:       item.UpdatedAt(),
	}
}

func (m *StockMapper) ToDomain(row *StockItemRow) (*stock.StockItem, error) {
	if row == nil {
		return nil, nil
	}

	sku, err := valueobject.NewSKU(row.SKU)
	if err != nil {
		return nil, err
	}

	return stock.RestoreStockItem(sku, row.Available, row.Reserved, row.MinimumQuantity, row.UpdatedAt), nil
}
