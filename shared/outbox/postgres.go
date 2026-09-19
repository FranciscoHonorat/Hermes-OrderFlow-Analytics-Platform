package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX é satisfeito tanto por *pgxpool.Pool quanto por pgx.Tx, permitindo
// que o mesmo repository seja usado fora de uma transação (relay/consulta)
// ou dentro de uma (unit of work da aplicação).
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

var _ Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	q DBTX
}

// NewPostgresRepository cria o repository do outbox a partir de qualquer
// DBTX: um *pgxpool.Pool (uso fora de transação) ou um pgx.Tx (dentro de
// uma transação de unit of work).
func NewPostgresRepository(q DBTX) *PostgresRepository {
	return &PostgresRepository{q: q}
}

func (r *PostgresRepository) SaveEvents(ctx context.Context, evts []events.DomainEvent) error {
	if len(evts) == 0 {
		return nil
	}

	query := `
		INSERT INTO outbox (id, aggregate_id, type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5);
	`

	for _, evt := range evts {
		payload, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("failed to marshal domain event: %w", err)
		}

		id := uuid.New()

		aggregateID := evt.AggregateId()
		eventType := evt.EventName()

		_, err = r.q.Exec(ctx, query,
			id,
			aggregateID,
			eventType,
			payload,
			time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert event into outbox: %w", err)
		}
	}

	return nil
}

func (r *PostgresRepository) FetchUnprocessed(ctx context.Context, limit int) ([]Row, error) {
	query := `
		SELECT id, aggregate_id, type, payload, created_at, processed_at
		FROM outbox
		WHERE processed_at IS NULL
		ORDER BY created_at
		LIMIT $1;
	`

	dbRows, err := r.q.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unprocessed outbox rows: %w", err)
	}
	defer dbRows.Close()

	rows := make([]Row, 0)
	for dbRows.Next() {
		var row Row
		if err := dbRows.Scan(&row.ID, &row.AggregateID, &row.Type, &row.Payload, &row.CreatedAt, &row.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func (r *PostgresRepository) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `UPDATE outbox SET processed_at = $1 WHERE id = $2;`
	now := time.Now().UTC()

	for _, id := range ids {
		if _, err := r.q.Exec(ctx, query, now, id); err != nil {
			return fmt.Errorf("failed to mark outbox event %s as processed: %w", id, err)
		}
	}

	return nil
}
