package output

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
)

type OutboxRepository interface {
	SaveEvents(ctx context.Context, events []event.DomainEvent) error
}
