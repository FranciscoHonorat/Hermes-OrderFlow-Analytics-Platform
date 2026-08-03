package postgres

import (
	"database/sql"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
)

type unitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) output.UnitOfWork {
	return &unitOfWork{db: db}
}

type repositoryProvider struct {
	tx *sql.Tx
}

func (p *repositoryProvider) OrderRepository() output.OrderRepository {
	return NewOrder
}
