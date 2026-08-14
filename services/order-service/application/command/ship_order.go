package command

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
)

type ShipOrderCommand struct {
	OrderID        string
	ShipmentID     string
	Carrier        string
	TrackingNumber string
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
			return domainErrors.ErrOrderNotFound
		}

		now := h.clock.Now()

		if err := order.Ship(cmd.ShipmentID, cmd.Carrier, cmd.TrackingNumber, now); err != nil {
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
