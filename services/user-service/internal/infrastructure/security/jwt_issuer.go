package security

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/auth"
)

type JWTIssuer struct {
	secret string
	expiry time.Duration
}

func NewJWTIssuer(secret string, expiry time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: secret, expiry: expiry}
}

func (i *JWTIssuer) Issue(userID, role string, now time.Time) (string, time.Time, error) {
	return auth.Issue(i.secret, userID, role, now, i.expiry)
}

func (i *JWTIssuer) Verify(tokenString string) (auth.Claims, error) {
	return auth.Verify(i.secret, tokenString)
}
