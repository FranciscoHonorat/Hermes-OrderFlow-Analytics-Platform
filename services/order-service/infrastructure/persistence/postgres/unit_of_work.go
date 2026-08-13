package postgres

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
	"github.com/jackc/pgx/v5"
)

type unitOfWork struct {
	db *DB
}

func NewUnitOfWork(db *DB) output.UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) Do(ctx context.Context, fn func(store output.RepositoryProvider) error) error {
	tx, err := u.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	provider := &repositoryProvider{
		tx:     tx,
		mapper: NewOrderMapper(),
	}

	if err := fn(provider); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type repositoryProvider struct {
	tx     pgx.Tx
	mapper *OrderMapper
}

func (r *repositoryProvider) OrderRepository() repository.OrderRepository {
	return NewOrderRepositoryFromTx(r.tx, r.mapper)
}

func (r *repositoryProvider) OutboxRepository() output.OutboxRepository {
	return NewOutboxRepositoryFromTx(r.tx)
}
