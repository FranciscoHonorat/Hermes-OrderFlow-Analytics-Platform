package user_test

import (
	"testing"
	"time"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedTime = time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

func newUser(t *testing.T, role string) *user.User {
	t.Helper()
	id := valueobject.NewUserIDMust(uuid.New())
	email := valueobject.NewEmailMust("user@example.com")
	r := valueobject.NewRoleMust(role)

	u, err := user.NewUser(id, email, "hashed-password", r, fixedTime)
	require.NoError(t, err)
	return u
}

func TestUser(t *testing.T) {
	validID := valueobject.NewUserIDMust(uuid.New())
	validEmail := valueobject.NewEmailMust("user@example.com")
	validRole := valueobject.NewRoleMust("user")

	t.Run("Test NewUser method", func(t *testing.T) {
		tests := []struct {
			name          string
			id            valueobject.UserID
			email         valueobject.Email
			passwordHash  string
			role          valueobject.Role
			expectedError error
		}{
			{"Valid user", validID, validEmail, "hashed-password", validRole, nil},
			{"Invalid ID", valueobject.UserID{}, validEmail, "hashed-password", validRole, domainErrors.ErrInvalidUserID},
			{"Invalid email", validID, valueobject.Email{}, "hashed-password", validRole, domainErrors.ErrInvalidEmail},
			{"Empty password hash", validID, validEmail, "", validRole, domainErrors.ErrInvalidPasswordHash},
			{"Invalid role", validID, validEmail, "hashed-password", valueobject.Role{}, domainErrors.ErrInvalidRole},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				u, err := user.NewUser(tt.id, tt.email, tt.passwordHash, tt.role, fixedTime)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.id, u.ID())
					assert.Equal(t, tt.email, u.Email())
					assert.True(t, u.Active())

					evts := u.PullEvents()
					require.Len(t, evts, 1)
					assert.Equal(t, "user.registered", evts[0].EventName())
				}
			})
		}
	})

	t.Run("Test ChangeRole method", func(t *testing.T) {
		tests := []struct {
			name          string
			newRole       string
			expectedError error
		}{
			{"Promote to admin", "admin", nil},
			{"Same role", "user", domainErrors.ErrRoleUnchanged},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				u := newUser(t, "user")
				u.ClearEvents()

				err := u.ChangeRole(valueobject.NewRoleMust(tt.newRole), fixedTime)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.Equal(t, "user", u.Role().String(), "state must be unchanged on failure")
					assert.Empty(t, u.DomainEvents(), "no event must be emitted on failure")
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.newRole, u.Role().String())

					evts := u.PullEvents()
					require.Len(t, evts, 1)
					assert.Equal(t, "user.role_changed", evts[0].EventName())
				}
			})
		}
	})

	t.Run("Test Deactivate method", func(t *testing.T) {
		t.Run("Valid deactivation", func(t *testing.T) {
			u := newUser(t, "user")
			u.ClearEvents()

			require.NoError(t, u.Deactivate(fixedTime))
			assert.False(t, u.Active())

			evts := u.PullEvents()
			require.Len(t, evts, 1)
			assert.Equal(t, "user.deactivated", evts[0].EventName())
		})

		t.Run("Already inactive", func(t *testing.T) {
			u := newUser(t, "user")
			require.NoError(t, u.Deactivate(fixedTime))
			u.ClearEvents()

			err := u.Deactivate(fixedTime)
			assert.ErrorIs(t, err, domainErrors.ErrUserAlreadyInactive)
			assert.Empty(t, u.DomainEvents())
		})
	})

	t.Run("Test Activate method", func(t *testing.T) {
		t.Run("Valid activation", func(t *testing.T) {
			u := newUser(t, "user")
			require.NoError(t, u.Deactivate(fixedTime))
			u.ClearEvents()

			require.NoError(t, u.Activate(fixedTime))
			assert.True(t, u.Active())

			evts := u.PullEvents()
			require.Len(t, evts, 1)
			assert.Equal(t, "user.activated", evts[0].EventName())
		})

		t.Run("Already active", func(t *testing.T) {
			u := newUser(t, "user")
			u.ClearEvents()

			err := u.Activate(fixedTime)
			assert.ErrorIs(t, err, domainErrors.ErrUserAlreadyActive)
			assert.Empty(t, u.DomainEvents())
		})
	})

	t.Run("Test ChangePassword method", func(t *testing.T) {
		tests := []struct {
			name          string
			newHash       string
			expectedError error
		}{
			{"Valid change", "new-hashed-password", nil},
			{"Same hash", "hashed-password", domainErrors.ErrPasswordUnchanged},
			{"Empty hash", "", domainErrors.ErrInvalidPasswordHash},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				u := newUser(t, "user")
				u.ClearEvents()

				err := u.ChangePassword(tt.newHash, fixedTime)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.Equal(t, "hashed-password", u.PasswordHash())
					assert.Empty(t, u.DomainEvents())
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.newHash, u.PasswordHash())

					evts := u.PullEvents()
					require.Len(t, evts, 1)
					assert.Equal(t, "user.password_changed", evts[0].EventName())
				}
			})
		}
	})
}

func TestRestoreUser(t *testing.T) {
	t.Run("Test RestoreUser method", func(t *testing.T) {
		id := valueobject.NewUserIDMust(uuid.New())
		email := valueobject.NewEmailMust("user@example.com")
		role := valueobject.NewRoleMust("admin")

		u := user.RestoreUser(id, email, "hashed-password", role, false, fixedTime, fixedTime)

		assert.Equal(t, id, u.ID())
		assert.Equal(t, email, u.Email())
		assert.Equal(t, "hashed-password", u.PasswordHash())
		assert.Equal(t, role, u.Role())
		assert.False(t, u.Active())
		assert.Empty(t, u.DomainEvents(), "restoring from persistence must not emit events")
	})
}
