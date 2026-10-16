package security

import "github.com/FranciscoHonorat/ordemflow/shared/auth"

type JWTVerifier struct {
	secret string
}

func NewJWTVerifier(secret string) *JWTVerifier {
	return &JWTVerifier{secret: secret}
}

func (v *JWTVerifier) Verify(token string) (auth.Claims, error) {
	return auth.Verify(v.secret, token)
}
