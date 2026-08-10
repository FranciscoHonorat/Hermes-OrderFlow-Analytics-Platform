package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
)

type ShipOrderCommand struct {
	OrderID string
}

type ShipOrderResult struct {
	OrderID string
}

type ShipOrderHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewShipOrderHandler(uow output.UnitOfWork, clock output.Clock) *ShipOrderHandler {
	return &ShipOrderHandler{
		uow:   uow,
		clock: clock,
	}
}

func (h *ShipOrderHandler) Handle(ctx context.Context, cmd ShipOrderCommand) error {
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
			return fmt.Errorf("order with ID %s not found", orderID.String())
		}

		now := h.clock.Now()

		if err := order.Ship(now); err != nil {
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
