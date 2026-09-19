package handler

import (
	"errors"
	"net/http"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/apperrors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	registerUseCase input.RegisterUserUseCase
	loginUseCase    input.LoginUseCase
}

func NewAuthHandler(register input.RegisterUserUseCase, login input.LoginUseCase) *AuthHandler {
	return &AuthHandler{registerUseCase: register, loginUseCase: login}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req input.RegisterUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	if err := h.registerUseCase.Execute(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req input.LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}

	result, err := h.loginUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, apperrors.ErrAccountDeactivated) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
