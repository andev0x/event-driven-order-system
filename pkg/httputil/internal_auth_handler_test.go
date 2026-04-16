package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestInternalAuthHandler_IssueToken(t *testing.T) {
	t.Run("returns unauthorized when header key is missing", func(t *testing.T) {
		handler := NewInternalAuthHandler("jwt-secret", "internal-key", "issuer", time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/internal/auth/token", nil)
		w := httptest.NewRecorder()

		handler.IssueToken(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("returns bad request for invalid ttl", func(t *testing.T) {
		handler := NewInternalAuthHandler("jwt-secret", "internal-key", "issuer", time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/internal/auth/token", strings.NewReader(`{"ttl_seconds":999999}`))
		req.Header.Set("X-Internal-Auth-Key", "internal-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.IssueToken(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns bad request for unknown field", func(t *testing.T) {
		handler := NewInternalAuthHandler("jwt-secret", "internal-key", "issuer", time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/internal/auth/token", strings.NewReader(`{"unknown":"value"}`))
		req.Header.Set("X-Internal-Auth-Key", "internal-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.IssueToken(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("issues signed token with defaults", func(t *testing.T) {
		handler := NewInternalAuthHandler("jwt-secret", "internal-key", "issuer", time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/internal/auth/token", strings.NewReader(`{}`))
		req.Header.Set("X-Internal-Auth-Key", "internal-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.IssueToken(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
		}

		var resp TokenResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.AccessToken == "" {
			t.Fatal("expected access token")
		}
		if resp.TokenType != "Bearer" {
			t.Errorf("token_type = %q, want %q", resp.TokenType, "Bearer")
		}
		if resp.ExpiresIn != 3600 {
			t.Errorf("expires_in = %d, want %d", resp.ExpiresIn, 3600)
		}

		token, err := jwt.Parse(resp.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte("jwt-secret"), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("expected jwt map claims")
		}

		if claims["sub"] != "internal-cli" {
			t.Errorf("sub = %v, want %q", claims["sub"], "internal-cli")
		}
		if claims["iss"] != "issuer" {
			t.Errorf("iss = %v, want %q", claims["iss"], "issuer")
		}
	})

	t.Run("caps default ttl to max allowed", func(t *testing.T) {
		handler := NewInternalAuthHandler("jwt-secret", "internal-key", "issuer", 48*time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/internal/auth/token", strings.NewReader(`{}`))
		req.Header.Set("X-Internal-Auth-Key", "internal-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.IssueToken(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
		}

		var resp TokenResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ExpiresIn != 86400 {
			t.Errorf("expires_in = %d, want %d", resp.ExpiresIn, 86400)
		}
	})
}
