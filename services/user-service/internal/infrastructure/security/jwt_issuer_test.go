package security_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTIssuer(t *testing.T) {
	t.Run("Test Issue and Verify round trip", func(t *testing.T) {
		issuer := security.NewJWTIssuer("test-secret", time.Hour)
		now := time.Now()

		token, expiresAt, err := issuer.Issue("user-1", "admin", now)
		require.NoError(t, err)
		assert.WithinDuration(t, now.Add(time.Hour), expiresAt, time.Second)

		claims, err := issuer.Verify(token)
		require.NoError(t, err)
		assert.Equal(t, "user-1", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("Test Verify rejects tampered secret", func(t *testing.T) {
		issuer := security.NewJWTIssuer("test-secret", time.Hour)
		other := security.NewJWTIssuer("other-secret", time.Hour)

		token, _, err := issuer.Issue("user-1", "admin", time.Now())
		require.NoError(t, err)

		_, err = other.Verify(token)
		assert.ErrorIs(t, err, security.ErrInvalidToken)
	})

	t.Run("Test Verify rejects expired token", func(t *testing.T) {
		issuer := security.NewJWTIssuer("test-secret", -time.Hour)

		token, _, err := issuer.Issue("user-1", "admin", time.Now())
		require.NoError(t, err)

		_, err = issuer.Verify(token)
		assert.ErrorIs(t, err, security.ErrInvalidToken)
	})

	t.Run("Test Verify rejects malformed token", func(t *testing.T) {
		issuer := security.NewJWTIssuer("test-secret", time.Hour)

		_, err := issuer.Verify("not-a-token")
		assert.ErrorIs(t, err, security.ErrInvalidToken)
	})
}
