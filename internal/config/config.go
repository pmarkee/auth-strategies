package config

import (
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"os"
)

type Config struct {
	Port int `yaml:"port"`
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

	log.Info().Int("port", cfg.Port).Msg("")
	return cfg
}
