package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
)

type DeactivateUserCommand struct {
	UserID string
}

type DeactivateUserHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewDeactivateUserHandler(uow output.UnitOfWork, clock output.Clock) *DeactivateUserHandler {
	return &DeactivateUserHandler{uow: uow, clock: clock}
}

func (h *DeactivateUserHandler) Handle(ctx context.Context, cmd DeactivateUserCommand) error {
	id, err := valueobject.ParseUserID(cmd.UserID)
	if err != nil {
		return err
	}

	now := h.clock.Now()

	return h.uow.Do(ctx, func(store output.RepositoryProvider) error {
		u, err := store.UserRepository().FindByID(ctx, id)
		if err != nil {
			return err
		}

		if err := u.Deactivate(now); err != nil {
			return err
		}

		if err := store.UserRepository().Save(ctx, u); err != nil {
			return fmt.Errorf("failed to save user: %w", err)
		}

		if evts := u.PullEvents(); len(evts) > 0 {
			if err := store.OutboxRepository().SaveEvents(ctx, evts); err != nil {
				return err
			}
		}

		return nil
	})
}

func (h *DeactivateUserHandler) Execute(ctx context.Context, in input.DeactivateUserInput) error {
	return h.Handle(ctx, DeactivateUserCommand{UserID: in.UserID})
}
