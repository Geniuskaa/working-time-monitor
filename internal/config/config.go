package config

import (
	"github.com/MicahParks/keyfunc"
	"github.com/jessevdk/go-flags"
	"os"
)

type App struct {
	Port string `env:"PORT" default:"7001"`
	Host string `env:"HOST" default:"0.0.0.0"`
}

type DB struct {
	MigrationsSourceURL string
	Hostname            string `env:"HOST"`
	Port                int    `env:"PORT"`
	Username            string `env:"USER"`
	Password            string `env:"PASSWORD"`
	DatabaseName        string `env:"DATABASE"`
}

type Keycloak struct {
	BasePath  string `env:"HOST"`
	Realm     string `env:"REALM"`
	PublicKey []byte
	JWK       *keyfunc.JWKS
}

type Config struct {
	App      App      `env-namespace:"APP"`
	DB       DB       `env-namespace:"PG"`
	Keycloak Keycloak `env-namespace:"KEYCLOAK"`
}

func Parse() (*Config, error) {
	var config Config
	p := flags.NewParser(&config, flags.HelpFlag|flags.PassDoubleDash)

	_, err := p.ParseArgs(os.Args)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
