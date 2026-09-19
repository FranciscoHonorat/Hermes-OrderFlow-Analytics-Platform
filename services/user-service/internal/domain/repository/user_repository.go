package repository

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
)

type UserRepository interface {
	Save(ctx context.Context, u *user.User) error
	FindByID(ctx context.Context, id valueobject.UserID) (*user.User, error)
	FindByEmail(ctx context.Context, email valueobject.Email) (*user.User, error)
}
