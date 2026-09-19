package security_test

import (
	"testing"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher(t *testing.T) {
	t.Run("Test Hash and Compare round trip", func(t *testing.T) {
		hasher := security.NewBcryptHasher(bcrypt.MinCost)

		hash, err := hasher.Hash("secret")
		require.NoError(t, err)
		assert.NotEqual(t, "secret", hash)

		assert.NoError(t, hasher.Compare(hash, "secret"))
		assert.Error(t, hasher.Compare(hash, "wrong"))
	})

	t.Run("Test NewBcryptHasher defaults invalid cost", func(t *testing.T) {
		hasher := security.NewBcryptHasher(0)

		hash, err := hasher.Hash("secret")
		require.NoError(t, err)
		assert.NoError(t, hasher.Compare(hash, "secret"))
	})
}
