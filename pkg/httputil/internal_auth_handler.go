package httputil

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultTokenSubject = "internal-cli"
	maxTokenTTLSeconds  = 24 * 60 * 60
)

type tokenRequest struct {
	Subject    string `json:"subject"`
	TTLSeconds int64  `json:"ttl_seconds"`
}

// TokenResponse represents an internal auth token response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// InternalAuthHandler issues JWT tokens for internal automation and tooling.
type InternalAuthHandler struct {
	secret      []byte
	internalKey string
	issuer      string
	defaultTTL  time.Duration
}

// NewInternalAuthHandler creates a token issuing handler secured by X-Internal-Auth-Key.
func NewInternalAuthHandler(jwtSecret, internalKey, issuer string, defaultTTL time.Duration) *InternalAuthHandler {
	if defaultTTL <= 0 {
		defaultTTL = time.Hour
	}
	maxTTL := time.Duration(maxTokenTTLSeconds) * time.Second
	if defaultTTL > maxTTL {
		defaultTTL = maxTTL
	}

	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		issuer = "internal-auth"
	}

	return &InternalAuthHandler{
		secret:      []byte(strings.TrimSpace(jwtSecret)),
		internalKey: strings.TrimSpace(internalKey),
		issuer:      issuer,
		defaultTTL:  defaultTTL,
	}
}

// IssueToken handles POST /internal/auth/token.
func (h *InternalAuthHandler) IssueToken(w http.ResponseWriter, r *http.Request) {
	if len(h.secret) == 0 {
		RespondError(w, http.StatusInternalServerError, "jwt secret not configured")
		return
	}

	if h.internalKey == "" {
		RespondError(w, http.StatusInternalServerError, "internal auth key not configured")
		return
	}

	providedKey := strings.TrimSpace(r.Header.Get("X-Internal-Auth-Key"))
	if subtle.ConstantTimeCompare([]byte(providedKey), []byte(h.internalKey)) != 1 {
		RespondError(w, http.StatusUnauthorized, "unauthorized internal auth request")
		return
	}

	req, ok := parseTokenRequest(w, r)
	if !ok {
		return
	}

	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		subject = defaultTokenSubject
	}

	ttl, ok := resolveTokenTTL(req.TTLSeconds, h.defaultTTL)
	if !ok {
		RespondError(w, http.StatusBadRequest, "ttl_seconds must be between 1 and 86400")
		return
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub": subject,
		"iss": h.issuer,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secret)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	RespondJSON(w, http.StatusOK, TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(ttl.Seconds()),
	})
}

func parseTokenRequest(w http.ResponseWriter, r *http.Request) (*tokenRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req tokenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)
	if err != nil && !errors.Is(err, io.EOF) {
		RespondError(w, http.StatusBadRequest, "invalid request payload")
		return nil, false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		RespondError(w, http.StatusBadRequest, "request body must contain only one JSON object")
		return nil, false
	}

	return &req, true
}

func resolveTokenTTL(ttlSeconds int64, defaultTTL time.Duration) (time.Duration, bool) {
	if ttlSeconds == 0 {
		return defaultTTL, true
	}

	if ttlSeconds < 1 || ttlSeconds > maxTokenTTLSeconds {
		return 0, false
	}

	return time.Duration(ttlSeconds) * time.Second, true
}
