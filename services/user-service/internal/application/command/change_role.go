package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
)

type ChangeRoleCommand struct {
	UserID string
	Role   string
}

type ChangeRoleHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewChangeRoleHandler(uow output.UnitOfWork, clock output.Clock) *ChangeRoleHandler {
	return &ChangeRoleHandler{uow: uow, clock: clock}
}

func (h *ChangeRoleHandler) Handle(ctx context.Context, cmd ChangeRoleCommand) error {
	id, err := valueobject.ParseUserID(cmd.UserID)
	if err != nil {
		return err
	}

	role, err := valueobject.NewRole(cmd.Role)
	if err != nil {
		return err
	}

	now := h.clock.Now()

	return h.uow.Do(ctx, func(store output.RepositoryProvider) error {
		u, err := store.UserRepository().FindByID(ctx, id)
		if err != nil {
			return err
		}

		if err := u.ChangeRole(role, now); err != nil {
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

func (h *ChangeRoleHandler) Execute(ctx context.Context, in input.ChangeRoleInput) error {
	return h.Handle(ctx, ChangeRoleCommand{UserID: in.UserID, Role: in.Role})
}
