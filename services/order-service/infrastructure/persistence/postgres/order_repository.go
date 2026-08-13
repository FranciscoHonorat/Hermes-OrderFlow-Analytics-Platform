package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
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

type OrderQueries struct {
	pool DBTX
}

var _ input.OrderQueries = (*OrderQueries)(nil)

func NewOrderQueries(db *DB) *OrderQueries {
	return &OrderQueries{pool: db.Pool}
}

func (q *OrderQueries) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*input.OrderDTO, error) {
	query := `
		SELECT id, customer_id, total_cents, status, created_at, updated_at
		FROM orders
		WHERE id = $1;
	`

	var (
		dbOrderID    string
		dbCustomerID string
		totalCents   int64
		status       string
		createdAt    time.Time
		updatedAt    time.Time
	)

	if err := q.pool.QueryRow(ctx, query, orderID.String()).Scan(
		&dbOrderID,
		&dbCustomerID,
		&totalCents,
		&status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, fmt.Errorf("query get order by id failed: %w", err)
	}

	orderUUID, err := uuid.Parse(dbOrderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order id from database: %w", err)
	}

	customerUUID, err := uuid.Parse(dbCustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer id from database: %w", err)
	}

	return &input.OrderDTO{
		ID:         orderUUID,
		CustomerID: customerUUID,
		Status:     status,
		Total:      totalCents,
		Currency:   "USD",
		Items:      []input.ItemDTO{},
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (q *OrderQueries) ListOrders(ctx context.Context, customerID *uuid.UUID, limit, offset int) ([]input.OrderDTO, error) {
	query := `
		SELECT id, customer_id, total_cents, status, created_at, updated_at
		FROM orders
	`
	args := []any{}

	if customerID != nil {
		query += ` WHERE customer_id = $1 `
		args = append(args, customerID.String())
	}

	query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2;`
	args = append(args, limit, offset)

	dbRows, err := q.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list orders failed: %w", err)
	}
	defer dbRows.Close()

	rows := make([]input.OrderDTO, 0)
	for dbRows.Next() {
		var (
			dbOrderID    string
			dbCustomerID string
			totalCents   int64
			status       string
			createdAt    time.Time
			updatedAt    time.Time
		)

		if err := dbRows.Scan(&dbOrderID, &dbCustomerID, &totalCents, &status, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}

		orderUUID, err := uuid.Parse(dbOrderID)
		if err != nil {
			return nil, fmt.Errorf("invalid order id from database: %w", err)
		}

		customerUUID, err := uuid.Parse(dbCustomerID)
		if err != nil {
			return nil, fmt.Errorf("invalid customer id from database: %w", err)
		}

		rows = append(rows, input.OrderDTO{
			ID:         orderUUID,
			CustomerID: customerUUID,
			Status:     status,
			Total:      totalCents,
			Currency:   "USD",
			Items:      []input.ItemDTO{},
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		})
	}

	return rows, nil
}
