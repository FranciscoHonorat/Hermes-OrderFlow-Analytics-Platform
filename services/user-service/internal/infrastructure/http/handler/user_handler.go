package handler

import (
	"net/http"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userQueries       input.UserQueries
	changeRoleUseCase input.ChangeRoleUseCase
	activateUseCase   input.ActivateUserUseCase
	deactivateUseCase input.DeactivateUserUseCase
}

func NewUserHandler(
	queries input.UserQueries,
	changeRole input.ChangeRoleUseCase,
	activate input.ActivateUserUseCase,
	deactivate input.DeactivateUserUseCase,
) *UserHandler {
	return &UserHandler{
		userQueries:       queries,
		changeRoleUseCase: changeRole,
		activateUseCase:   activate,
		deactivateUseCase: deactivate,
	}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	value, exists := c.Get(middleware.ClaimsKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication"})
		return
	}

	claims, ok := value.(middleware.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication"})
		return
	}

	dto, err := h.userQueries.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	dto, err := h.userQueries.GetUserByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	items, err := h.userQueries.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *UserHandler) ChangeRole(c *gin.Context) {
	var req input.ChangeRoleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed or invalid JSON body"})
		return
	}
	req.UserID = c.Param("id")

	if err := h.changeRoleUseCase.Execute(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role changed successfully"})
}

func (h *UserHandler) ActivateUser(c *gin.Context) {
	if err := h.activateUseCase.Execute(c.Request.Context(), input.ActivateUserInput{UserID: c.Param("id")}); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}

func (h *UserHandler) DeactivateUser(c *gin.Context) {
	if err := h.deactivateUseCase.Execute(c.Request.Context(), input.DeactivateUserInput{UserID: c.Param("id")}); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}
