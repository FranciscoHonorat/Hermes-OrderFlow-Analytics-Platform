package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		claims         *middleware.Claims
		allowedRoles   []string
		expectedStatus int
	}{
		{"Allowed role", &middleware.Claims{Role: "admin"}, []string{"admin"}, http.StatusOK},
		{"Disallowed role", &middleware.Claims{Role: "user"}, []string{"admin"}, http.StatusForbidden},
		{"No claims in context", nil, []string{"admin"}, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/admin", func(c *gin.Context) {
				if tt.claims != nil {
					c.Set(middleware.ClaimsKey, *tt.claims)
				}
				c.Next()
			}, middleware.RequireRole(tt.allowedRoles...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
