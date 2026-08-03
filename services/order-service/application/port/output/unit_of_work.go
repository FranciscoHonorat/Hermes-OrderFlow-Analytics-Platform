package output

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
)

type RepositoryProvider interface {
	OrderRepository() repository.OrderRepository
	OutboxRepository() OutboxRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(store RepositoryProvider) error) error
}
