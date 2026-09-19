package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmail(t *testing.T) {
	t.Run("Test NewEmail method", func(t *testing.T) {
		tests := []struct {
			name          string
			value         string
			expectedValue string
			expectedError error
		}{
			{"Valid email", "user@example.com", "user@example.com", nil},
			{"Normalizes to lowercase", "User@Example.COM", "user@example.com", nil},
			{"Trims surrounding whitespace", "  user@example.com  ", "user@example.com", nil},
			{"Invalid email (empty)", "", "", domainErrors.ErrInvalidEmail},
			{"Invalid email (only whitespace)", "   ", "", domainErrors.ErrInvalidEmail},
			{"Invalid email (malformed)", "not-an-email", "", domainErrors.ErrInvalidEmail},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				email, err := valueobject.NewEmail(tt.value)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.True(t, email.IsZero())
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedValue, email.String())
					assert.False(t, email.IsZero())
				}
			})
		}
	})

	t.Run("Test Equal method", func(t *testing.T) {
		a := valueobject.NewEmailMust("user@example.com")
		b := valueobject.NewEmailMust("USER@example.com")
		c := valueobject.NewEmailMust("other@example.com")

		assert.True(t, a.Equal(b), "comparison must be case-insensitive via normalization")
		assert.False(t, a.Equal(c))
	})

	t.Run("Test MarshalJSON and UnmarshalJSON", func(t *testing.T) {
		original := valueobject.NewEmailMust("user@example.com")

		data, err := original.MarshalJSON()
		require.NoError(t, err)
		assert.JSONEq(t, `"user@example.com"`, string(data))

		var roundTripped valueobject.Email
		require.NoError(t, roundTripped.UnmarshalJSON(data))
		assert.True(t, original.Equal(roundTripped))
	})

	t.Run("Test UnmarshalJSON with invalid value", func(t *testing.T) {
		var email valueobject.Email
		err := email.UnmarshalJSON([]byte(`"not-an-email"`))
		assert.ErrorIs(t, err, domainErrors.ErrInvalidEmail)
	})
}
