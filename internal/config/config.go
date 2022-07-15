package config

import (
	"fmt"
	"github.com/MicahParks/keyfunc"
	"github.com/ilyakaznacheev/cleanenv"
)

type App struct {
	Port string `yaml:"port" env:"PORT" env-default:"7001"`
	Host string `yaml:"host" env:"HOST" env-default:"0.0.0.0"`
}

type DB struct {
	MigrationsSourceURL string
	Hostname            string `yaml:"hostname" env:"PG_HOST"`
	Port                int    `yaml:"port" env:"PG_PORT"`
	Username            string `yaml:"username" env:"PG_USER"`
	Password            string `yaml:"password" env:"PG_PASSWORD"`
	DatabaseName        string `yaml:"databaseName" env:"PG_DATABASE"`
}

type Keycloak struct {
	BasePath  string `yaml:"base-path" env:"KEYCLOAK_HOST"`
	Realm     string `yaml:"realm" env:"KEYCLOAK_REALM"`
	PublicKey []byte
	JWK       *keyfunc.JWKS
}

type Config struct {
	App      App      `yaml:"app"`
	DB       DB       `yaml:"db"`
	Keycloak Keycloak `yaml:"keycloak"`
}

func NewConfig() (cfg *Config, err error) {
	cfg = &Config{}
	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		description, err := cleanenv.GetDescription(cfg, nil)
		fmt.Println(description)
		return nil, err
	}
	return cfg, nil
}
