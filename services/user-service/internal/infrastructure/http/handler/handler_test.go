package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/apperrors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/handler"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type mockRegisterUseCase struct {
	err      error
	gotInput input.RegisterUserInput
}

func (m *mockRegisterUseCase) Execute(ctx context.Context, in input.RegisterUserInput) error {
	m.gotInput = in
	return m.err
}

type mockLoginUseCase struct {
	result input.LoginResult
	err    error
}

func (m *mockLoginUseCase) Execute(ctx context.Context, in input.LoginInput) (input.LoginResult, error) {
	return m.result, m.err
}

type mockChangeRoleUseCase struct {
	err      error
	gotInput input.ChangeRoleInput
}

func (m *mockChangeRoleUseCase) Execute(ctx context.Context, in input.ChangeRoleInput) error {
	m.gotInput = in
	return m.err
}

type mockActivateUseCase struct {
	err error
}

func (m *mockActivateUseCase) Execute(ctx context.Context, in input.ActivateUserInput) error {
	return m.err
}

type mockDeactivateUseCase struct {
	err error
}

func (m *mockDeactivateUseCase) Execute(ctx context.Context, in input.DeactivateUserInput) error {
	return m.err
}

type mockUserQueries struct {
	dto     *input.UserDTO
	getErr  error
	list    []input.UserDTO
	listErr error
}

func (m *mockUserQueries) GetUserByID(ctx context.Context, id string) (*input.UserDTO, error) {
	return m.dto, m.getErr
}

func (m *mockUserQueries) ListUsers(ctx context.Context) ([]input.UserDTO, error) {
	return m.list, m.listErr
}

func newJSONContext(t *testing.T, method, path string, body any) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return w, c
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("valid request registers the user", func(t *testing.T) {
		register := &mockRegisterUseCase{}
		h := handler.NewAuthHandler(register, &mockLoginUseCase{})

		w, c := newJSONContext(t, "POST", "/auth/register", input.RegisterUserInput{Email: "user@example.com", Password: "secret"})
		h.Register(c)

		require.Equal(t, 201, w.Code)
		require.Equal(t, "user@example.com", register.gotInput.Email)
	})

	t.Run("malformed body returns 400", func(t *testing.T) {
		h := handler.NewAuthHandler(&mockRegisterUseCase{}, &mockLoginUseCase{})

		w, c := newJSONContext(t, "POST", "/auth/register", nil)
		h.Register(c)

		require.Equal(t, 400, w.Code)
	})

	t.Run("use case error returns 422", func(t *testing.T) {
		register := &mockRegisterUseCase{err: errors.New("email already registered")}
		h := handler.NewAuthHandler(register, &mockLoginUseCase{})

		w, c := newJSONContext(t, "POST", "/auth/register", input.RegisterUserInput{Email: "user@example.com", Password: "secret"})
		h.Register(c)

		require.Equal(t, 422, w.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("valid credentials returns a token", func(t *testing.T) {
		login := &mockLoginUseCase{result: input.LoginResult{AccessToken: "token", ExpiresAt: time.Now()}}
		h := handler.NewAuthHandler(&mockRegisterUseCase{}, login)

		w, c := newJSONContext(t, "POST", "/auth/login", input.LoginInput{Email: "user@example.com", Password: "secret"})
		h.Login(c)

		require.Equal(t, 200, w.Code)
	})

	t.Run("invalid credentials returns 401", func(t *testing.T) {
		login := &mockLoginUseCase{err: apperrors.ErrInvalidCredentials}
		h := handler.NewAuthHandler(&mockRegisterUseCase{}, login)

		w, c := newJSONContext(t, "POST", "/auth/login", input.LoginInput{Email: "user@example.com", Password: "secret"})
		h.Login(c)

		require.Equal(t, 401, w.Code)
	})

	t.Run("deactivated account returns 403", func(t *testing.T) {
		login := &mockLoginUseCase{err: apperrors.ErrAccountDeactivated}
		h := handler.NewAuthHandler(&mockRegisterUseCase{}, login)

		w, c := newJSONContext(t, "POST", "/auth/login", input.LoginInput{Email: "user@example.com", Password: "secret"})
		h.Login(c)

		require.Equal(t, 403, w.Code)
	})
}

func TestUserHandler_GetMe(t *testing.T) {
	t.Run("returns the authenticated user's profile", func(t *testing.T) {
		queries := &mockUserQueries{dto: &input.UserDTO{ID: "user-1", Email: "user@example.com"}}
		h := handler.NewUserHandler(queries, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "GET", "/users/me", nil)
		c.Set(middleware.ClaimsKey, middleware.Claims{UserID: "user-1", Role: "user"})
		h.GetMe(c)

		require.Equal(t, 200, w.Code)
	})

	t.Run("missing claims returns 401", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserQueries{}, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "GET", "/users/me", nil)
		h.GetMe(c)

		require.Equal(t, 401, w.Code)
	})
}

func TestUserHandler_GetUserByID(t *testing.T) {
	t.Run("found returns 200", func(t *testing.T) {
		queries := &mockUserQueries{dto: &input.UserDTO{ID: "user-1"}}
		h := handler.NewUserHandler(queries, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "GET", "/users/user-1", nil)
		c.Params = gin.Params{{Key: "id", Value: "user-1"}}
		h.GetUserByID(c)

		require.Equal(t, 200, w.Code)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		queries := &mockUserQueries{getErr: errors.New("not found")}
		h := handler.NewUserHandler(queries, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "GET", "/users/missing", nil)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		h.GetUserByID(c)

		require.Equal(t, 404, w.Code)
	})
}

func TestUserHandler_ListUsers(t *testing.T) {
	t.Run("success returns 200", func(t *testing.T) {
		queries := &mockUserQueries{list: []input.UserDTO{{ID: "user-1"}}}
		h := handler.NewUserHandler(queries, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "GET", "/users", nil)
		h.ListUsers(c)

		require.Equal(t, 200, w.Code)
	})

	t.Run("query error returns 500", func(t *testing.T) {
		queries := &mockUserQueries{listErr: errors.New("db error")}
		h := handler.NewUserHandler(queries, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "GET", "/users", nil)
		h.ListUsers(c)

		require.Equal(t, 500, w.Code)
	})
}

func TestUserHandler_ChangeRole(t *testing.T) {
	t.Run("valid request changes role", func(t *testing.T) {
		changeRole := &mockChangeRoleUseCase{}
		h := handler.NewUserHandler(&mockUserQueries{}, changeRole, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "PATCH", "/users/user-1/role", input.ChangeRoleInput{Role: "admin"})
		c.Params = gin.Params{{Key: "id", Value: "user-1"}}
		h.ChangeRole(c)

		require.Equal(t, 200, w.Code)
		require.Equal(t, "user-1", changeRole.gotInput.UserID)
		require.Equal(t, "admin", changeRole.gotInput.Role)
	})

	t.Run("use case error returns 422", func(t *testing.T) {
		changeRole := &mockChangeRoleUseCase{err: errors.New("role unchanged")}
		h := handler.NewUserHandler(&mockUserQueries{}, changeRole, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "PATCH", "/users/user-1/role", input.ChangeRoleInput{Role: "admin"})
		c.Params = gin.Params{{Key: "id", Value: "user-1"}}
		h.ChangeRole(c)

		require.Equal(t, 422, w.Code)
	})
}

func TestUserHandler_ActivateUser(t *testing.T) {
	t.Run("valid request activates the user", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserQueries{}, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "POST", "/users/user-1/activate", nil)
		c.Params = gin.Params{{Key: "id", Value: "user-1"}}
		h.ActivateUser(c)

		require.Equal(t, 200, w.Code)
	})
}

func TestUserHandler_DeactivateUser(t *testing.T) {
	t.Run("valid request deactivates the user", func(t *testing.T) {
		h := handler.NewUserHandler(&mockUserQueries{}, &mockChangeRoleUseCase{}, &mockActivateUseCase{}, &mockDeactivateUseCase{})

		w, c := newJSONContext(t, "POST", "/users/user-1/deactivate", nil)
		c.Params = gin.Params{{Key: "id", Value: "user-1"}}
		h.DeactivateUser(c)

		require.Equal(t, 200, w.Code)
	})
}
