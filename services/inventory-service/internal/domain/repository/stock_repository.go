package repository

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/stock"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
)

type StockRepository interface {
	Save(ctx context.Context, item *stock.StockItem) error
	FindBySKU(ctx context.Context, sku valueobject.SKU) (*stock.StockItem, error)
}
