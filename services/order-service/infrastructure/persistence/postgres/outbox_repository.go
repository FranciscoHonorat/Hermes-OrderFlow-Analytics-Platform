package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OutboxRow struct {
	ID          uuid.UUID  `db:"id"`
	AggregateID string     `db:"aggregate_id"`
	Type        string     `db:"type"`
	Payload     []byte     `db:"payload"`
	CreatedAt   time.Time  `db:"created_at"`
	ProcessedAt *time.Time `db:"processed_at"`
}

type OutboxRepository struct {
	q DBTX
}

func NewOutboxRepository(db *DB) output.OutboxRepository {
	return &OutboxRepository{q: db.Pool}
}

func NewOutboxRepositoryFromTx(tx pgx.Tx) output.OutboxRepository {
	return &OutboxRepository{q: tx}
}

// SaveEvents implementa com perfeição o contrato da interface output.OutboxRepository
func (r *OutboxRepository) SaveEvents(ctx context.Context, events []event.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	query := `
		INSERT INTO outbox (id, aggregate_id, type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5);
	`

	for _, evt := range events {
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
