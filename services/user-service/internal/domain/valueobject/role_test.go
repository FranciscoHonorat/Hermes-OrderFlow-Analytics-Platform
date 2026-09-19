package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRole(t *testing.T) {
	t.Run("Test NewRole method", func(t *testing.T) {
		tests := []struct {
			name          string
			value         string
			expectedAdmin bool
			expectedError error
		}{
			{"Valid admin role", "admin", true, nil},
			{"Valid user role", "user", false, nil},
			{"Invalid role", "superuser", false, domainErrors.ErrInvalidRole},
			{"Invalid role (empty)", "", false, domainErrors.ErrInvalidRole},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				role, err := valueobject.NewRole(tt.value)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.True(t, role.IsZero())
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.value, role.String())
					assert.Equal(t, tt.expectedAdmin, role.IsAdmin())
				}
			})
		}
	})

	t.Run("Test Equal method", func(t *testing.T) {
		a := valueobject.NewRoleMust("admin")
		b := valueobject.NewRoleMust("admin")
		c := valueobject.NewRoleMust("user")

		assert.True(t, a.Equal(b))
		assert.False(t, a.Equal(c))
	})

	t.Run("Test MarshalJSON and UnmarshalJSON", func(t *testing.T) {
		original := valueobject.NewRoleMust("admin")

		data, err := original.MarshalJSON()
		require.NoError(t, err)
		assert.JSONEq(t, `"admin"`, string(data))

		var roundTripped valueobject.Role
		require.NoError(t, roundTripped.UnmarshalJSON(data))
		assert.True(t, original.Equal(roundTripped))
	})

	t.Run("Test UnmarshalJSON with invalid value", func(t *testing.T) {
		var role valueobject.Role
		err := role.UnmarshalJSON([]byte(`"superuser"`))
		assert.ErrorIs(t, err, domainErrors.ErrInvalidRole)
	})
}
