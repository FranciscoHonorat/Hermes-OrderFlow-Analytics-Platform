package output

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
)

type OutboxRepository = outbox.Repository

type RepositoryProvider interface {
	UserRepository() repository.UserRepository
	OutboxRepository() OutboxRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(store RepositoryProvider) error) error
}
