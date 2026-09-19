package postgres

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/stock"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type StockRepository struct {
	q      DBTX
	mapper *StockMapper
}

func NewStockRepository(db *DB, mapper *StockMapper) repository.StockRepository {
	return &StockRepository{q: db.Pool, mapper: mapper}
}

func NewStockRepositoryFromTx(tx pgx.Tx, mapper *StockMapper) repository.StockRepository {
	return &StockRepository{q: tx, mapper: mapper}
}

func (r *StockRepository) Save(ctx context.Context, item *stock.StockItem) error {
	row := r.mapper.ToPersistence(item)

	query := `
		INSERT INTO stock_items (sku, available, reserved, minimum_quantity, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sku) DO UPDATE SET
			available = EXCLUDED.available,
			reserved = EXCLUDED.reserved,
			minimum_quantity = EXCLUDED.minimum_quantity,
			updated_at = EXCLUDED.updated_at;
	`

	_, err := r.q.Exec(ctx, query,
		row.SKU,
		row.Available,
		row.Reserved,
		row.MinimumQuantity,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save stock item to postgres: %w", err)
	}

	return nil
}

func (r *StockRepository) FindBySKU(ctx context.Context, sku valueobject.SKU) (*stock.StockItem, error) {
	query := `
		SELECT sku, available, reserved, minimum_quantity, updated_at
		FROM stock_items
		WHERE sku = $1;
	`

	var row StockItemRow

	err := r.q.QueryRow(ctx, query, sku.String()).Scan(
		&row.SKU,
		&row.Available,
		&row.Reserved,
		&row.MinimumQuantity,
		&row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find stock item %s: %w", sku.String(), err)
	}

	return r.mapper.ToDomain(&row)
}

type StockQueries struct {
	pool DBTX
}

var _ input.StockQueries = (*StockQueries)(nil)

func NewStockQueries(db *DB) *StockQueries {
	return &StockQueries{pool: db.Pool}
}

func (q *StockQueries) GetStockBySKU(ctx context.Context, sku string) (*input.StockDTO, error) {
	query := `
		SELECT sku, available, reserved, minimum_quantity, updated_at
		FROM stock_items
		WHERE sku = $1;
	`

	var dto input.StockDTO
	if err := q.pool.QueryRow(ctx, query, sku).Scan(
		&dto.SKU,
		&dto.Available,
		&dto.Reserved,
		&dto.MinimumQuantity,
		&dto.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("query get stock by sku failed: %w", err)
	}

	return &dto, nil
}

func (q *StockQueries) ListLowStock(ctx context.Context) ([]input.StockDTO, error) {
	query := `
		SELECT sku, available, reserved, minimum_quantity, updated_at
		FROM stock_items
		WHERE available < minimum_quantity
		ORDER BY sku;
	`

	dbRows, err := q.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query list low stock failed: %w", err)
	}
	defer dbRows.Close()

	rows := make([]input.StockDTO, 0)
	for dbRows.Next() {
		var dto input.StockDTO
		if err := dbRows.Scan(&dto.SKU, &dto.Available, &dto.Reserved, &dto.MinimumQuantity, &dto.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan stock item row: %w", err)
		}
		rows = append(rows, dto)
	}

	return rows, nil
}
