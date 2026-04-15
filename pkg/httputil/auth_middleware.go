package httputil

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type claimsContextKey string

const jwtClaimsKey claimsContextKey = "jwt_claims"

// JWTMiddleware validates bearer tokens signed with HMAC SHA-256.
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	secret = strings.TrimSpace(secret)
	secretBytes := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(secretBytes) == 0 {
				RespondError(w, http.StatusInternalServerError, "authentication not configured")
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				RespondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			tokenString, ok := extractBearerToken(authHeader)
			if !ok {
				RespondError(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return secretBytes, nil
			},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithExpirationRequired(),
			)
			if err != nil || !token.Valid {
				RespondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				RespondError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), jwtClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// JWTClaimsFromContext extracts JWT claims added by JWTMiddleware.
func JWTClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(jwtClaimsKey).(jwt.MapClaims)
	return claims, ok
}

func extractBearerToken(authHeader string) (string, bool) {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}
