package postgres

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
)

type UserRow struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	Active       bool      `db:"active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (m *UserMapper) ToPersistence(u *user.User) *UserRow {
	if u == nil {
		return nil
	}

	return &UserRow{
		ID:           u.ID().String(),
		Email:        u.Email().String(),
		PasswordHash: u.PasswordHash(),
		Role:         u.Role().String(),
		Active:       u.Active(),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
	}
}

func (m *UserMapper) ToDomain(row *UserRow) (*user.User, error) {
	if row == nil {
		return nil, nil
	}

	id, err := valueobject.ParseUserID(row.ID)
	if err != nil {
		return nil, err
	}

	email, err := valueobject.NewEmail(row.Email)
	if err != nil {
		return nil, err
	}

	role, err := valueobject.NewRole(row.Role)
	if err != nil {
		return nil, err
	}

	return user.RestoreUser(id, email, row.PasswordHash, role, row.Active, row.CreatedAt, row.UpdatedAt), nil
}
