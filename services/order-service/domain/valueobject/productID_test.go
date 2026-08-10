package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	product "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductID(t *testing.T) {
	t.Run("Test for NewProductID", func(t *testing.T) {
		tests := []struct {
			name          string
			id            uuid.UUID
			expectedError error
		}{
			{
				name:          "Valid ProductID",
				id:            uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				expectedError: nil,
			},
			{
				name:          "Invalid ProductID (Nil UUID)",
				id:            uuid.Nil,
				expectedError: domainErrors.ErrInvalidProductID,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				pid, err := product.NewProductID(tt.id)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					require.NoError(t, err)
					require.NotNil(t, pid)
					assert.Equal(t, tt.id, pid.ID())
				}
			})
		}
	})

	t.Run("Test for NewProductIDMust", func(t *testing.T) {
		validID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

		assert.NotPanics(t, func() {
			pid := product.NewProductIDMust(validID)
			assert.Equal(t, validID, pid.ID())
		})

		assert.Panics(t, func() {
			product.NewProductIDMust(uuid.Nil)
		})
	})

	t.Run("Test for String method", func(t *testing.T) {
		id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		pid, err := product.NewProductID(id)
		require.NoError(t, err)

		assert.Equal(t, id.String(), pid.String())
	})

	t.Run("Test for Equal method", func(t *testing.T) {
		id1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		id2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

		pid1 := product.NewProductIDMust(id1)
		pid2 := product.NewProductIDMust(id1)
		pid3 := product.NewProductIDMust(id2)

		assert.True(t, pid1.Equal(pid2), "Expected pid1 to be equal to pid2")
		assert.False(t, pid1.Equal(pid3), "Expected pid1 to not be equal to pid3")
	})

	t.Run("Test for MarshalJSON", func(t *testing.T) {
		id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		tests := []struct {
			name     string
			pid      product.ProductID
			expected string
		}{
			{"Valid ID", product.NewProductIDMust(id), `{"id":"` + id.String() + `"}`},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonData, err := tt.pid.MarshalJSON()
				require.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(jsonData))
			})
		}
	})

	t.Run("Test for UnmarshalJSON", func(t *testing.T) {
		id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		tests := []struct {
			name          string
			inputJSON     string
			expectedID    uuid.UUID
			expectedError error
		}{
			{"Valid JSON", `{"id":"` + id.String() + `"}`, id, nil},
			{"Invalid JSON ID", `{"id":"00000000-0000-0000-0000-000000000000"}`, uuid.Nil, domainErrors.ErrInvalidProductID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var unmarshalledPID product.ProductID
				err := unmarshalledPID.UnmarshalJSON([]byte(tt.inputJSON))

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedID, unmarshalledPID.ID())
				}
			})
		}
	})
}
