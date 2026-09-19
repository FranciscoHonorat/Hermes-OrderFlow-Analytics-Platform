package command

import (
	"context"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/google/uuid"
)

type RegisterUserCommand struct {
	Email    string
	Password string
	Role     string
}

type RegisterUserHandler struct {
	uow    output.UnitOfWork
	clock  output.Clock
	hasher output.PasswordHasher
}

func NewRegisterUserHandler(uow output.UnitOfWork, clock output.Clock, hasher output.PasswordHasher) *RegisterUserHandler {
	return &RegisterUserHandler{uow: uow, clock: clock, hasher: hasher}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) error {
	email, err := valueobject.NewEmail(cmd.Email)
	if err != nil {
		return err
	}

	role, err := valueobject.NewRole(cmd.Role)
	if err != nil {
		return err
	}

	hash, err := h.hasher.Hash(cmd.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := h.clock.Now()

	newUser, err := user.NewUser(valueobject.NewUserIDMust(uuid.New()), email, hash, role, now)
	if err != nil {
		return err
	}

	return h.uow.Do(ctx, func(store output.RepositoryProvider) error {
		if err := store.UserRepository().Save(ctx, newUser); err != nil {
			return err
		}

		if evts := newUser.PullEvents(); len(evts) > 0 {
			if err := store.OutboxRepository().SaveEvents(ctx, evts); err != nil {
				return err
			}
		}

		return nil
	})
}

func (h *RegisterUserHandler) Execute(ctx context.Context, in input.RegisterUserInput) error {
	return h.Handle(ctx, RegisterUserCommand{Email: in.Email, Password: in.Password, Role: valueobject.RoleNameUser})
}
