package auth_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	t.Run("Test Issue and Verify round trip", func(t *testing.T) {
		now := time.Now()

		token, expiresAt, err := auth.Issue("test-secret", "user-1", "admin", now, time.Hour)
		require.NoError(t, err)
		assert.WithinDuration(t, now.Add(time.Hour), expiresAt, time.Second)

		claims, err := auth.Verify("test-secret", token)
		require.NoError(t, err)
		assert.Equal(t, "user-1", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("Test Verify rejects tampered secret", func(t *testing.T) {
		token, _, err := auth.Issue("test-secret", "user-1", "admin", time.Now(), time.Hour)
		require.NoError(t, err)

		_, err = auth.Verify("other-secret", token)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	t.Run("Test Verify rejects expired token", func(t *testing.T) {
		token, _, err := auth.Issue("test-secret", "user-1", "admin", time.Now(), -time.Hour)
		require.NoError(t, err)

		_, err = auth.Verify("test-secret", token)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	t.Run("Test Verify rejects malformed token", func(t *testing.T) {
		_, err := auth.Verify("test-secret", "not-a-token")
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})
}
