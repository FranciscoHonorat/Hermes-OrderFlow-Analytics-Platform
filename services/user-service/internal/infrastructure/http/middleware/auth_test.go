package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeVerifier struct {
	claims middleware.Claims
	err    error
}

func (f fakeVerifier) Verify(token string) (middleware.Claims, error) {
	return f.claims, f.err
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		header         string
		verifier       middleware.TokenVerifier
		expectedStatus int
	}{
		{"Valid token", "Bearer good-token", fakeVerifier{claims: middleware.Claims{UserID: "user-1", Role: "user"}}, http.StatusOK},
		{"Missing header", "", fakeVerifier{}, http.StatusUnauthorized},
		{"Malformed header", "good-token", fakeVerifier{}, http.StatusUnauthorized},
		{"Invalid token", "Bearer bad-token", fakeVerifier{err: errors.New("invalid")}, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.Auth(tt.verifier))
			router.GET("/protected", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
