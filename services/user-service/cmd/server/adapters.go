package main

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/middleware"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/security"
)

type tokenVerifierAdapter struct {
	issuer *security.JWTIssuer
}

func (a *tokenVerifierAdapter) Verify(token string) (middleware.Claims, error) {
	claims, err := a.issuer.Verify(token)
	if err != nil {
		return middleware.Claims{}, err
	}
	return middleware.Claims{UserID: claims.UserID, Role: claims.Role}, nil
}

type registerUserUseCaseAdapter struct {
	handler *command.RegisterUserHandler
}

func (a *registerUserUseCaseAdapter) Execute(ctx context.Context, in input.RegisterUserInput) error {
	return a.handler.Execute(ctx, in)
}

type changeRoleUseCaseAdapter struct {
	handler *command.ChangeRoleHandler
}

func (a *changeRoleUseCaseAdapter) Execute(ctx context.Context, in input.ChangeRoleInput) error {
	return a.handler.Execute(ctx, in)
}

type activateUserUseCaseAdapter struct {
	handler *command.ActivateUserHandler
}

func (a *activateUserUseCaseAdapter) Execute(ctx context.Context, in input.ActivateUserInput) error {
	return a.handler.Execute(ctx, in)
}

type deactivateUserUseCaseAdapter struct {
	handler *command.DeactivateUserHandler
}

func (a *deactivateUserUseCaseAdapter) Execute(ctx context.Context, in input.DeactivateUserInput) error {
	return a.handler.Execute(ctx, in)
}

type loginUseCaseAdapter struct {
	handler *command.LoginHandler
}

func (a *loginUseCaseAdapter) Execute(ctx context.Context, in input.LoginInput) (input.LoginResult, error) {
	return a.handler.Execute(ctx, in)
}
