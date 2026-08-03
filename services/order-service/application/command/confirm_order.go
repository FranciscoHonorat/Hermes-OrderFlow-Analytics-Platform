package command

import (
	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
)

type ConfirmOrderCommand struct {
	OrderID string
}

type ConfirmOrderResult struct {
	OrderID string
	Status  string
}

type ConfirmOrderHandler struct {
	orders    repository.OrderRepository
	uow       output.UnitOfWork
	publisher output.EventPublisher
}
