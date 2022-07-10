package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"strings"
)

type authMiddleware struct {
	cfg *config.Config
	db  *postgres.Db
	log *zap.SugaredLogger
}

func NewMiddleware(cfg *config.Config, db *postgres.Db, log *zap.SugaredLogger) *authMiddleware {
	return &authMiddleware{cfg: cfg, db: db, log: log}
}

const keycloakUsernameClaimKey string = "preferred_username"

func (m *authMiddleware) Middleware(next http.Handler) http.Handler {

	f := func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Authorization token missing", http.StatusUnauthorized)
			return
		}

		token = m.extractBearerToken(token)
		jwtToken, err := jwt.Parse(token, m.cfg.Keycloak.JWK.Keyfunc)
		if err != nil || !jwtToken.Valid {
			http.Error(w, fmt.Sprintf("Invalid token %s", token), http.StatusUnauthorized)
			return
		}
		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		username, ok := claims[keycloakUsernameClaimKey].(string)
		if !ok {
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
