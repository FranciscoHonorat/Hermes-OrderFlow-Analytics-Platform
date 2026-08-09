package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	address "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress(t *testing.T) {
	t.Run("Tests for NewAddress", func(t *testing.T) {
		tests := []struct {
			name           string
			cep            string
			street         string
			neighborhood   string
			number         int64
			referencePoint string
			complement     string
			expectedError  error
		}{
			{"valid Address", "12345678", "Street", "Neighborhood", 2, "reference", "complement", nil},
			{"Invalid CEP", "123456", "Street", "Neighborhood", 2, "reference", "complement", domainErrors.ErrInvalidCEP},
			{"Invalid Street", "12345678", "", "Neighborhood", 2, "reference", "complement", domainErrors.ErrFieldEmpty},
			{"Invalid Neighborhood", "12345678", "Street", "", 2, "reference", "complement", domainErrors.ErrFieldEmpty},
			{"Invalid Number", "12345678", "Street", "Neighborhood", -1, "reference", "complement", domainErrors.ErrInvalidNumber},
			{"Invalid ReferencePoint", "12345678", "Street", "Neighborhood", 2, "", "complement", domainErrors.ErrFieldEmpty},
			{"Invalid Complement", "12345678", "Street", "Neighborhood", 2, "reference", "", domainErrors.ErrFieldEmpty},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				a, err := address.NewAddress(tt.cep, tt.street, tt.neighborhood, tt.number, tt.referencePoint, tt.complement)
				assert.ErrorIs(t, err, tt.expectedError)
				if tt.expectedError == nil {
					require.NotNil(t, a)
					assert.Equal(t, tt.cep, a.Cep())
					assert.Equal(t, tt.street, a.Street())
					assert.Equal(t, tt.neighborhood, a.Neighborhood())
					assert.Equal(t, tt.number, a.Number())
					assert.Equal(t, tt.referencePoint, a.ReferencePoint())
					assert.Equal(t, tt.complement, a.Complement())
				}
			})
		}
	})

	t.Run("Test for MarshalJSON", func(t *testing.T) {
		tests := []struct {
			name     string
			a        address.Address
			expected string
		}{
			{
				name:     "valid Address",
				a:        address.NewAddressMust("12345678", "Street", "Neighborhood", 2, "reference", "complement"),
				expected: `{"cep":"12345678","street":"Street","neighborhood":"Neighborhood","number":2,"referencePoint":"reference","complement":"complement"}`,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonData, err := tt.a.MarshalJSON()
				require.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(jsonData))
			})
		}
	})

	t.Run("Test for UnmarshalJSON", func(t *testing.T) {
		tests := []struct {
			name            string
			jsonData        string
			expectedAddress address.Address
			expectedError   error
		}{
			{"Valid JSON", `{"cep":"12345678","street":"Street","neighborhood":"Neighborhood","number":2,"referencePoint":"reference","complement":"complement"}`, address.NewAddressMust("12345678", "Street", "Neighborhood", 2, "reference", "complement"), nil},
			{"Invalid Street", `{"cep":"12345678","street":"","neighborhood":"Neighborhood","number":2,"referencePoint":"reference","complement":"complement"}`, address.Address{}, domainErrors.ErrFieldEmpty},
			{"Invalid Neighborhood", `{"cep":"12345678","street":"Street","neighborhood":"","number":2,"referencePoint":"reference","complement":"complement"}`, address.Address{}, domainErrors.ErrFieldEmpty},
			{"Invalid Number", `{"cep":"12345678","street":"Street","neighborhood":"Neighborhood","number":-1,"referencePoint":"reference","complement":"complement"}`, address.Address{}, domainErrors.ErrInvalidNumber},
			{"Invalid ReferencePoint", `{"cep":"12345678","street":"Street","neighborhood":"Neighborhood","number":2,"referencePoint":"","complement":"complement"}`, address.Address{}, domainErrors.ErrFieldEmpty},
			{"Invalid Complement", `{"cep":"12345678","street":"Street","neighborhood":"Neighborhood","number":2,"referencePoint":"reference","complement":""}`, address.Address{}, domainErrors.ErrFieldEmpty},
		}
		for _, tt := range tests {
			var a address.Address
			err := a.UnmarshalJSON([]byte(tt.jsonData))
			assert.ErrorIs(t, err, tt.expectedError)
		}
	})
}
