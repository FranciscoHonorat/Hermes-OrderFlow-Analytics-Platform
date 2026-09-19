package output

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
)

// OutboxRepository é a porta usada pela aplicação; a definição real do
// contrato vive em shared/outbox, reaproveitada por todo bounded context
// que grava eventos de domínio na tabela outbox.
type OutboxRepository = outbox.Repository

type RepositoryProvider interface {
	StockRepository() repository.StockRepository
	OutboxRepository() OutboxRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(store RepositoryProvider) error) error
}
