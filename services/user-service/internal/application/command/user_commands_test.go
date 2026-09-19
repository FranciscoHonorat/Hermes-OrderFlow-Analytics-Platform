package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/apperrors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/shared/events"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type MockClock struct {
	NowTime time.Time
}

func (m *MockClock) Now() time.Time {
	return m.NowTime
}

type MockPasswordHasher struct {
	HashPrefix string
	CompareErr error
}

func (m *MockPasswordHasher) Hash(plain string) (string, error) {
	prefix := m.HashPrefix
	if prefix == "" {
		prefix = "hashed:"
	}
	return prefix + plain, nil
}

func (m *MockPasswordHasher) Compare(hash, plain string) error {
	return m.CompareErr
}

type MockTokenIssuer struct {
	Token     string
	ExpiresAt time.Time
	Err       error
}

func (m *MockTokenIssuer) Issue(userID, role string, now time.Time) (string, time.Time, error) {
	if m.Err != nil {
		return "", time.Time{}, m.Err
	}
	return m.Token, m.ExpiresAt, nil
}

type MockUserRepository struct {
	ByID    *user.User
	ByEmail *user.User
	Saved   *user.User
	SaveErr error
}

func (m *MockUserRepository) Save(ctx context.Context, u *user.User) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.Saved = u
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id valueobject.UserID) (*user.User, error) {
	if m.ByID == nil {
		return nil, errors.New("user not found")
	}
	return m.ByID, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email valueobject.Email) (*user.User, error) {
	if m.ByEmail == nil {
		return nil, errors.New("user not found")
	}
	return m.ByEmail, nil
}

type MockOutboxRepository struct {
	SavedEvents []events.DomainEvent
}

func (m *MockOutboxRepository) SaveEvents(ctx context.Context, evts []events.DomainEvent) error {
	m.SavedEvents = append(m.SavedEvents, evts...)
	return nil
}

func (m *MockOutboxRepository) FetchUnprocessed(ctx context.Context, limit int) ([]outbox.Row, error) {
	return nil, nil
}

func (m *MockOutboxRepository) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	return nil
}

type MockRepositoryProvider struct {
	userRepo   *MockUserRepository
	outboxRepo *MockOutboxRepository
}

func (m *MockRepositoryProvider) UserRepository() repository.UserRepository {
	return m.userRepo
}

func (m *MockRepositoryProvider) OutboxRepository() output.OutboxRepository {
	return m.outboxRepo
}

type MockUnitOfWork struct {
	provider *MockRepositoryProvider
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(store output.RepositoryProvider) error) error {
	return fn(m.provider)
}

var fixedTime = time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

func newTestUser(t *testing.T, role string, active bool) *user.User {
	t.Helper()
	id := valueobject.NewUserIDMust(uuid.New())
	email := valueobject.NewEmailMust("user@example.com")
	u, err := user.NewUser(id, email, "hashed:secret", valueobject.NewRoleMust(role), fixedTime)
	require.NoError(t, err)
	u.ClearEvents()

	if !active {
		require.NoError(t, u.Deactivate(fixedTime))
		u.ClearEvents()
	}

	return u
}

func TestRegisterUserHandler_Handle(t *testing.T) {
	t.Run("registers a new user and drains events to outbox", func(t *testing.T) {
		userRepo := &MockUserRepository{}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewRegisterUserHandler(uow, &MockClock{NowTime: fixedTime}, &MockPasswordHasher{})

		err := handler.Handle(context.Background(), command.RegisterUserCommand{
			Email:    "user@example.com",
			Password: "secret",
			Role:     "user",
		})
		require.NoError(t, err)

		require.NotNil(t, userRepo.Saved)
		require.Equal(t, "hashed:secret", userRepo.Saved.PasswordHash())
		require.Len(t, outboxRepo.SavedEvents, 1)
		require.Equal(t, "user.registered", outboxRepo.SavedEvents[0].EventName())
	})

	t.Run("returns error for invalid email without touching the repository", func(t *testing.T) {
		userRepo := &MockUserRepository{}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewRegisterUserHandler(uow, &MockClock{NowTime: fixedTime}, &MockPasswordHasher{})

		err := handler.Handle(context.Background(), command.RegisterUserCommand{
			Email:    "not-an-email",
			Password: "secret",
			Role:     "user",
		})
		require.Error(t, err)
		require.Nil(t, userRepo.Saved)
	})

	t.Run("propagates unique-email error from repository", func(t *testing.T) {
		userRepo := &MockUserRepository{SaveErr: errors.New("email already registered")}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewRegisterUserHandler(uow, &MockClock{NowTime: fixedTime}, &MockPasswordHasher{})

		err := handler.Handle(context.Background(), command.RegisterUserCommand{
			Email:    "user@example.com",
			Password: "secret",
			Role:     "user",
		})
		require.Error(t, err)
		require.Empty(t, outboxRepo.SavedEvents)
	})

	t.Run("Execute always forces role to user", func(t *testing.T) {
		userRepo := &MockUserRepository{}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewRegisterUserHandler(uow, &MockClock{NowTime: fixedTime}, &MockPasswordHasher{})

		err := handler.Execute(context.Background(), input.RegisterUserInput{Email: "user@example.com", Password: "secret"})
		require.NoError(t, err)
		require.Equal(t, "user", userRepo.Saved.Role().String())
	})
}

func TestChangeRoleHandler_Handle(t *testing.T) {
	t.Run("promotes a user to admin", func(t *testing.T) {
		existing := newTestUser(t, "user", true)
		userRepo := &MockUserRepository{ByID: existing}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewChangeRoleHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ChangeRoleCommand{UserID: existing.ID().String(), Role: "admin"})
		require.NoError(t, err)

		require.Equal(t, "admin", userRepo.Saved.Role().String())
		require.Len(t, outboxRepo.SavedEvents, 1)
		require.Equal(t, "user.role_changed", outboxRepo.SavedEvents[0].EventName())
	})

	t.Run("returns error when role is unchanged", func(t *testing.T) {
		existing := newTestUser(t, "user", true)
		userRepo := &MockUserRepository{ByID: existing}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewChangeRoleHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ChangeRoleCommand{UserID: existing.ID().String(), Role: "user"})
		require.Error(t, err)
		require.Nil(t, userRepo.Saved)
	})
}

func TestActivateUserHandler_Handle(t *testing.T) {
	t.Run("activates an inactive user", func(t *testing.T) {
		existing := newTestUser(t, "user", false)
		userRepo := &MockUserRepository{ByID: existing}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewActivateUserHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ActivateUserCommand{UserID: existing.ID().String()})
		require.NoError(t, err)
		require.True(t, userRepo.Saved.Active())
		require.Len(t, outboxRepo.SavedEvents, 1)
		require.Equal(t, "user.activated", outboxRepo.SavedEvents[0].EventName())
	})
}

func TestDeactivateUserHandler_Handle(t *testing.T) {
	t.Run("deactivates an active user", func(t *testing.T) {
		existing := newTestUser(t, "user", true)
		userRepo := &MockUserRepository{ByID: existing}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{userRepo: userRepo, outboxRepo: outboxRepo}}

		handler := command.NewDeactivateUserHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.DeactivateUserCommand{UserID: existing.ID().String()})
		require.NoError(t, err)
		require.False(t, userRepo.Saved.Active())
		require.Len(t, outboxRepo.SavedEvents, 1)
		require.Equal(t, "user.deactivated", outboxRepo.SavedEvents[0].EventName())
	})
}

func TestLoginHandler_Handle(t *testing.T) {
	t.Run("issues a token for valid credentials", func(t *testing.T) {
		existing := newTestUser(t, "user", true)
		userRepo := &MockUserRepository{ByEmail: existing}
		issuer := &MockTokenIssuer{Token: "signed-token", ExpiresAt: fixedTime.Add(time.Hour)}

		handler := command.NewLoginHandler(userRepo, &MockPasswordHasher{}, issuer, &MockClock{NowTime: fixedTime})

		result, err := handler.Handle(context.Background(), command.LoginCommand{Email: "user@example.com", Password: "secret"})
		require.NoError(t, err)
		require.Equal(t, "signed-token", result.AccessToken)
	})

	t.Run("returns invalid credentials for unknown email", func(t *testing.T) {
		userRepo := &MockUserRepository{}
		issuer := &MockTokenIssuer{}

		handler := command.NewLoginHandler(userRepo, &MockPasswordHasher{}, issuer, &MockClock{NowTime: fixedTime})

		_, err := handler.Handle(context.Background(), command.LoginCommand{Email: "user@example.com", Password: "secret"})
		require.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
	})

	t.Run("returns invalid credentials for wrong password", func(t *testing.T) {
		existing := newTestUser(t, "user", true)
		userRepo := &MockUserRepository{ByEmail: existing}
		hasher := &MockPasswordHasher{CompareErr: errors.New("mismatch")}
		issuer := &MockTokenIssuer{}

		handler := command.NewLoginHandler(userRepo, hasher, issuer, &MockClock{NowTime: fixedTime})

		_, err := handler.Handle(context.Background(), command.LoginCommand{Email: "user@example.com", Password: "wrong"})
		require.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
	})

	t.Run("returns account deactivated for inactive user", func(t *testing.T) {
		existing := newTestUser(t, "user", false)
		userRepo := &MockUserRepository{ByEmail: existing}
		issuer := &MockTokenIssuer{}

		handler := command.NewLoginHandler(userRepo, &MockPasswordHasher{}, issuer, &MockClock{NowTime: fixedTime})

		_, err := handler.Handle(context.Background(), command.LoginCommand{Email: "user@example.com", Password: "secret"})
		require.ErrorIs(t, err, apperrors.ErrAccountDeactivated)
	})
}
