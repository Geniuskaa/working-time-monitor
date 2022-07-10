package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type App struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type DB struct {
	MigrationsSourceURL string `yaml:"migrationsSourceURL"`
	Hostname            string `yaml:"hostname"`
	Port                int    `yaml:"port"`
	Username            string `yaml:"username"`
	Password            string `yaml:"password"`
	DatabaseName        string `yaml:"databaseName"`
}

type Keycloak struct {
	Host         string `yaml:"host"`
	ClientId     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
	Realm        string `yaml:"realm"`
}

type Config struct {
	App      App      `yaml:"app"`
	DB       DB       `yaml:"db"`
	Keycloak Keycloak `yaml:"keycloak"`
}

func NewConfig(profile string) (cfg *Config, port, host string, err error) {
	cfg = &Config{}
	cfgPath := fmt.Sprintf("configs/%s.yml", profile)
	err = cleanenv.ReadConfig(cfgPath, cfg)
	if err != nil {
		return nil, "", "", err
	}
	port = cfg.App.Port
	host = cfg.App.Host

	return cfg, port, host, nil
}
