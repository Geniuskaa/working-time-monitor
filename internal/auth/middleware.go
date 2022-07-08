package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"strings"
)

type authMiddleware struct {
	keycloak *keycloak
	db       *postgres.Db
	log      *zap.SugaredLogger
}

func NewMiddleware(cfg *config.Config, db *postgres.Db, log *zap.SugaredLogger) *authMiddleware {
	return &authMiddleware{keycloak: newKeycloak(cfg), db: db, log: log}
}

func (m *authMiddleware) Authenticate(next http.Handler) http.Handler {

	f := func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Authorization token missing", http.StatusUnauthorized)
			return
		}

		token = m.extractBearerToken(token)
		if token == "" {
			http.Error(w, fmt.Sprintf("Invalid token %s", token), http.StatusUnauthorized)
			return
		}

		result, err := m.keycloak.gocloak.RetrospectToken(context.Background(), token, m.keycloak.clientId, m.keycloak.clientSecret, m.keycloak.realm)
		if !*result.Active {
			http.Error(w, "Account is not active", http.StatusUnauthorized)
			return
		}
		username, err := m.extractUsername(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token %s", token), http.StatusUnauthorized)
			return
		}
		user, err := m.db.GetUserPrincipalByUsername(context.Background(), username)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token %s", token), http.StatusUnauthorized)
			return
		}
		principal := UserPrincipal{Id: user.Id, Username: user.Username, Email: user.Email}
		r = r.WithContext(context.WithValue(r.Context(), userPrincipalContextKey, &principal))
		m.log.Infof("User %s authenticated with token %s", principal.Username, token)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (m *authMiddleware) extractBearerToken(token string) string {
	if !strings.HasPrefix(token, "Bearer ") {
		return ""
	}
	return strings.Replace(token, "Bearer ", "", 1)
}

const keycloakUsernameClaimKey string = "preferred_username"

func (m *authMiddleware) extractUsername(token string) (string, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	t, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", err
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok {
		return claims[keycloakUsernameClaimKey].(string), nil
	}
	return "", errors.New("error extracting username")
}
