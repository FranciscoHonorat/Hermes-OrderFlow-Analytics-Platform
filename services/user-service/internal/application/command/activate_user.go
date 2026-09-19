package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
)

type ActivateUserCommand struct {
	UserID string
}

type ActivateUserHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewActivateUserHandler(uow output.UnitOfWork, clock output.Clock) *ActivateUserHandler {
	return &ActivateUserHandler{uow: uow, clock: clock}
}

func (h *ActivateUserHandler) Handle(ctx context.Context, cmd ActivateUserCommand) error {
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

		if err := u.Activate(now); err != nil {
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

func (h *ActivateUserHandler) Execute(ctx context.Context, in input.ActivateUserInput) error {
	return h.Handle(ctx, ActivateUserCommand{UserID: in.UserID})
}
