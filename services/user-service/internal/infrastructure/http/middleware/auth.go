package middleware

import (
	"net/http"

	"github.com/FranciscoHonorat/ordemflow/shared/auth"
	"github.com/gin-gonic/gin"
)

const ClaimsKey = "auth_claims"

type TokenVerifier interface {
	Verify(token string) (auth.Claims, error)
}

func Auth(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "

		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed Authorization header"})
			return
		}

		claims, err := verifier.Verify(header[len(prefix):])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}
