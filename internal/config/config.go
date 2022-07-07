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

type Config struct {
	App App `yaml:"app"`
	DB  DB  `yaml:"db"`
}

func NewConfig(profile string) (*Config, error) {
	cfg := Config{}
	cfgPath := fmt.Sprintf("configs/%s.yml", profile)
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
