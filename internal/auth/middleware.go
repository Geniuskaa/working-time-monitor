package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"strings"
)

type authMiddleware struct {
	keycloak *keycloak
}

func NewMiddleware(cfg *config.Config) *authMiddleware {
	return &authMiddleware{keycloak: newKeycloak(cfg)}
}

func (m *authMiddleware) Authenticate(next http.Handler) http.Handler {

	f := func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("Authorization")

		if token == "" {
			http.Error(w, "Authorization token missing", http.StatusUnauthorized)
			return
		}

		token, err := m.extractBearerToken(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token %s", token), http.StatusUnauthorized)
			return
		}

		result, err := m.keycloak.gocloak.RetrospectToken(context.Background(), token, m.keycloak.clientId, m.keycloak.clientSecret, m.keycloak.realm)
		if !*result.Active {
			http.Error(w, "Account is not active", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (m *authMiddleware) extractBearerToken(token string) (string, error) {
	if !strings.HasPrefix(token, "Bearer ") {
		return "", errors.New("invalid token")
	}
	return strings.Replace(token, "Bearer ", "", 1), nil
}
