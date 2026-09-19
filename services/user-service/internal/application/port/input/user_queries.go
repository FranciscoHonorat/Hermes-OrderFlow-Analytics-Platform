package input

import (
	"context"
	"time"
)

type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserQueries interface {
	GetUserByID(ctx context.Context, id string) (*UserDTO, error)
	ListUsers(ctx context.Context) ([]UserDTO, error)
}
