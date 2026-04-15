package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTMiddleware(t *testing.T) {
	t.Run("returns unauthorized when header is missing", func(t *testing.T) {
		middleware := JWTMiddleware("test-secret")
		handlerCalled := false

		handler := middleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", w.Code, http.StatusUnauthorized)
		}
		if handlerCalled {
			t.Error("expected handler not to be called")
		}
	})

	t.Run("returns unauthorized when token is invalid", func(t *testing.T) {
		middleware := JWTMiddleware("test-secret")
		handlerCalled := false

		handler := middleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", w.Code, http.StatusUnauthorized)
		}
		if handlerCalled {
			t.Error("expected handler not to be called")
		}
	})

	t.Run("allows request with valid token", func(t *testing.T) {
		secret := "test-secret"
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		middleware := JWTMiddleware(secret)
		handlerCalled := false

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			claims, ok := JWTClaimsFromContext(r.Context())
			if !ok {
				t.Error("expected claims in context")
				return
			}

			if claims["sub"] != "user-123" {
				t.Errorf("claims[sub] = %v, want user-123", claims["sub"])
			}

			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("status = %v, want %v", w.Code, http.StatusNoContent)
		}
		if !handlerCalled {
			t.Error("expected handler to be called")
		}
	})
}
