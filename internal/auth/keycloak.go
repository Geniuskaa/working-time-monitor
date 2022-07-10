package auth

import (
	"github.com/Nerzal/gocloak/v11"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
)

type keycloak struct {
	gocloak      gocloak.GoCloak
	clientId     string
	clientSecret string
	realm        string
}

func newKeycloak(cfg *config.Config) *keycloak {
	return &keycloak{
		gocloak:      gocloak.NewClient(cfg.Keycloak.Host),
		clientId:     cfg.Keycloak.ClientId,
		clientSecret: cfg.Keycloak.ClientSecret,
		realm:        cfg.Keycloak.Realm,
	}
}
