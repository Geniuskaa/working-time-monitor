package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type App struct {
}

type DB struct {
}

type Config struct {
	App App `yaml:"app"`
	DB  DB  `yaml:"db"`
}

func NewConfig(profile string) (*Config, error) {
	cfg := &Config{}
	cfgPath := fmt.Sprintf("configs/%s.yml", profile)
	err := cleanenv.ReadConfig(cfgPath, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
