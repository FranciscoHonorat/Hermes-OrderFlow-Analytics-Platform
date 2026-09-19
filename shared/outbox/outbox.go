package outbox

import (
	"context"
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
	"github.com/google/uuid"
)

// Row é a representação de uma linha da tabela `outbox`, comum a todo
// bounded context que adota o padrão Transactional Outbox (ver ADR-003).
type Row struct {
	ID          uuid.UUID
	AggregateID string
	Type        string
	Payload     []byte
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// Repository é o contrato compartilhado do outbox. SaveEvents é usado pela
// aplicação, dentro da mesma transação que persiste o agregado.
// FetchUnprocessed e MarkProcessed são usados pelo relay (cdc-connector)
// que lê a tabela e publica os eventos no Kafka.
type Repository interface {
	SaveEvents(ctx context.Context, evts []events.DomainEvent) error
	FetchUnprocessed(ctx context.Context, limit int) ([]Row, error)
	MarkProcessed(ctx context.Context, ids []uuid.UUID) error
}
