package config

import (
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"os"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Db     DbConfig     `yaml:"db"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DbConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func ParseConfig() Config {
	cfgBytes, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read config.yaml")
	}

	var cfg Config
	if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal config.yaml")
	}

	log.Info().Msgf("read config: %+v", cfg)
	return cfg
}
