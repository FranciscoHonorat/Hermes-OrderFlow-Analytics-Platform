package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
)

type ReleaseStockCommand struct {
	SKU      string
	Quantity int
}

type ReleaseStockHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewReleaseStockHandler(uow output.UnitOfWork, clock output.Clock) *ReleaseStockHandler {
	return &ReleaseStockHandler{uow: uow, clock: clock}
}

func (h *ReleaseStockHandler) Handle(ctx context.Context, cmd ReleaseStockCommand) error {
	sku, err := valueobject.NewSKU(cmd.SKU)
	if err != nil {
		return err
	}

	now := h.clock.Now()

	return h.uow.Do(ctx, func(store output.RepositoryProvider) error {
		item, err := store.StockRepository().FindBySKU(ctx, sku)
		if err != nil {
			return err
		}

		if err := item.Release(cmd.Quantity, now); err != nil {
			return err
		}

		if err := store.StockRepository().Save(ctx, item); err != nil {
			return fmt.Errorf("failed to save stock item: %w", err)
		}

		if evts := item.PullEvents(); len(evts) > 0 {
			if err := store.OutboxRepository().SaveEvents(ctx, evts); err != nil {
				return err
			}
		}

		return nil
	})
}

func (h *ReleaseStockHandler) Execute(ctx context.Context, in input.ReleaseStockInput) error {
	return h.Handle(ctx, ReleaseStockCommand{SKU: in.SKU, Quantity: in.Quantity})
}
