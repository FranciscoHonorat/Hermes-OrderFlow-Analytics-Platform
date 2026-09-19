package command

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/apperrors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
)

type LoginCommand struct {
	Email    string
	Password string
}

type LoginHandler struct {
	users  repository.UserRepository
	hasher output.PasswordHasher
	issuer output.TokenIssuer
	clock  output.Clock
}

func NewLoginHandler(users repository.UserRepository, hasher output.PasswordHasher, issuer output.TokenIssuer, clock output.Clock) *LoginHandler {
	return &LoginHandler{users: users, hasher: hasher, issuer: issuer, clock: clock}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (input.LoginResult, error) {
	email, err := valueobject.NewEmail(cmd.Email)
	if err != nil {
		return input.LoginResult{}, apperrors.ErrInvalidCredentials
	}

	u, err := h.users.FindByEmail(ctx, email)
	if err != nil {
		return input.LoginResult{}, apperrors.ErrInvalidCredentials
	}

	if err := h.hasher.Compare(u.PasswordHash(), cmd.Password); err != nil {
		return input.LoginResult{}, apperrors.ErrInvalidCredentials
	}

	if !u.Active() {
		return input.LoginResult{}, apperrors.ErrAccountDeactivated
	}

	now := h.clock.Now()

	token, expiresAt, err := h.issuer.Issue(u.ID().String(), u.Role().String(), now)
	if err != nil {
		return input.LoginResult{}, err
	}

	return input.LoginResult{AccessToken: token, ExpiresAt: expiresAt}, nil
}

func (h *LoginHandler) Execute(ctx context.Context, in input.LoginInput) (input.LoginResult, error) {
	return h.Handle(ctx, LoginCommand{Email: in.Email, Password: in.Password})
}
