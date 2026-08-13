package postgres

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type OrderRepository struct {
	q      DBTX
	mapper *OrderMapper
}

func NewOrderRepository(db *DB, mapper *OrderMapper) repository.OrderRepository {
	return &OrderRepository{
		q:      db.Pool,
		mapper: mapper,
	}
}

func NewOrderRepositoryFromTx(tx pgx.Tx, mapper *OrderMapper) repository.OrderRepository {
	return &OrderRepository{
		q:      tx,
		mapper: mapper,
	}
}

func (r *OrderRepository) Save(ctx context.Context, order *entity.Order) error {
	row := r.mapper.ToPersistence(order)

	query := `
		INSERT INTO orders (id, customer_id, total_cents, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			total_cents = EXCLUDED.total_cents,
			updated_at = EXCLUDED.updated_at;
	`

	_, err := r.q.Exec(ctx, query,
		row.ID,
		row.CustomerID,
		row.TotalCents,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save order to postgres: %w", err)
	}

	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id valueobject.OrderID) (*entity.Order, error) {
	query := `
		SELECT id, customer_id, total_cents, status, created_at, updated_at 
		FROM orders 
		WHERE id = $1;
	`

	var row OrderRow

	err := r.q.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.CustomerID,
		&row.TotalCents,
		&row.Status,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find order %s: %w", id.String(), err)
	}

	domainOrder, err := r.mapper.ToDomain(&row)
	if err != nil {
		return nil, fmt.Errorf("failed to map order row to domain: %w", err)
	}

	return domainOrder, nil
}
