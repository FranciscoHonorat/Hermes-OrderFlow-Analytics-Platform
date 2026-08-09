package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
)

type CancelOrderCommand struct {
	OrderID string
}

type CancelOrderResult struct {
	OrderID string
}

type CancelOrderHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewCancelOrderHandler(uow output.UnitOfWork, clock output.Clock) *CancelOrderHandler {
	return &CancelOrderHandler{
		uow:   uow,
		clock: clock,
	}
}

func (h *CancelOrderHandler) Handle(ctx context.Context, cmd CancelOrderCommand) error {
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
		if err := order.Cancel(now, "canceled by user"); err != nil {
			return err
		}

		if err := store.OrderRepository().Save(ctx, order); err != nil {
			return err
		}

		if events := order.PullEvents(); len(events) > 0 {
			if err := store.OutboxRepository().SaveEvents(ctx, events); err != nil {
				return err
			}
		}

		return nil
	})
}
