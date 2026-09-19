package input

import (
	"context"
	"time"
)

type RegisterUserInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterUserUseCase interface {
	Execute(ctx context.Context, in RegisterUserInput) error
}

type ChangeRoleInput struct {
	UserID string `json:"-"`
	Role   string `json:"role" binding:"required"`
}

type ChangeRoleUseCase interface {
	Execute(ctx context.Context, in ChangeRoleInput) error
}

type ActivateUserInput struct {
	UserID string `json:"-"`
}

type ActivateUserUseCase interface {
	Execute(ctx context.Context, in ActivateUserInput) error
}

type DeactivateUserInput struct {
	UserID string `json:"-"`
}

type DeactivateUserUseCase interface {
	Execute(ctx context.Context, in DeactivateUserInput) error
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResult struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type LoginUseCase interface {
	Execute(ctx context.Context, in LoginInput) (LoginResult, error)
}
