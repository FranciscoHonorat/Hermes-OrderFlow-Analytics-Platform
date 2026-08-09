package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
)

type ConfirmOrderCommand struct {
	OrderID string
}

type ConfirmOrderResult struct {
	OrderID string
}

type ConfirmOrderHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewConfirmOrderHandler(uow output.UnitOfWork, clock output.Clock) *ConfirmOrderHandler {
	return &ConfirmOrderHandler{
		uow:   uow,
		clock: clock,
	}
}

func (h *ConfirmOrderHandler) Handle(ctx context.Context, cmd ConfirmOrderCommand) error {
	orderUUID, err := uuid.Parse(cmd.OrderID)
	if err != nil {
		return err
	}
	orderID, err := valueobject.NewOrderID(orderUUID)
	if err != nil {
		return err
	}

	return h.uow.Do(ctx, func(store output.RepositoryProvider) error {
		order, err := store.OrderRepository().FindByID(ctx, orderID)
		if err != nil {
			return err
		}
		if order == nil {
			return fmt.Errorf("order not found")
		}

		now := h.clock.Now()
		if err := order.Confirm(now); err != nil {
			return err
		}

		if err := store.OrderRepository().Save(ctx, order); err != nil {
			return err
		}

		if events := order.PullEvents(); len(events) > 0 {
			return store.OutboxRepository().SaveEvents(ctx, events)
		}

		return nil
	})
}
