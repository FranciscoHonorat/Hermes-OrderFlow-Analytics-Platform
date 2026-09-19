package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserID(t *testing.T) {
	t.Run("Test NewUserID method", func(t *testing.T) {
		valid := uuid.New()

		tests := []struct {
			name          string
			id            uuid.UUID
			expectedError error
		}{
			{"Valid UUID", valid, nil},
			{"Nil UUID", uuid.Nil, domainErrors.ErrInvalidUserID},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				userID, err := valueobject.NewUserID(tt.id)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.True(t, userID.IsZero())
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.id, userID.ID())
					assert.False(t, userID.IsZero())
				}
			})
		}
	})

	t.Run("Test ParseUserID method", func(t *testing.T) {
		valid := uuid.New()

		tests := []struct {
			name          string
			raw           string
			expectedError error
		}{
			{"Valid UUID string", valid.String(), nil},
			{"Malformed string", "not-a-uuid", domainErrors.ErrInvalidUserID},
			{"Nil UUID string", uuid.Nil.String(), domainErrors.ErrInvalidUserID},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				userID, err := valueobject.ParseUserID(tt.raw)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, valid, userID.ID())
				}
			})
		}
	})

	t.Run("Test Equal method", func(t *testing.T) {
		id := uuid.New()
		a := valueobject.NewUserIDMust(id)
		b := valueobject.NewUserIDMust(id)
		c := valueobject.NewUserIDMust(uuid.New())

		assert.True(t, a.Equal(b))
		assert.False(t, a.Equal(c))
	})

	t.Run("Test MarshalJSON and UnmarshalJSON", func(t *testing.T) {
		original := valueobject.NewUserIDMust(uuid.New())

		data, err := original.MarshalJSON()
		require.NoError(t, err)

		var roundTripped valueobject.UserID
		require.NoError(t, roundTripped.UnmarshalJSON(data))
		assert.True(t, original.Equal(roundTripped))
	})

	t.Run("Test UnmarshalJSON with invalid value", func(t *testing.T) {
		var userID valueobject.UserID
		err := userID.UnmarshalJSON([]byte(`""`))
		assert.ErrorIs(t, err, domainErrors.ErrInvalidUserID)
	})
}
