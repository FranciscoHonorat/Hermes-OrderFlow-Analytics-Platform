package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID string
	Role   string
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type JWTIssuer struct {
	secret []byte
	expiry time.Duration
}

func NewJWTIssuer(secret string, expiry time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), expiry: expiry}
}

func (i *JWTIssuer) Issue(userID, role string, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(i.expiry)

	claims := jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (i *JWTIssuer) Verify(tokenString string) (Claims, error) {
	var claims jwtClaims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return i.secret, nil
	})
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	return Claims{UserID: claims.Subject, Role: claims.Role}, nil
}
