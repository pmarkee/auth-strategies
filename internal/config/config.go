package config

import (
	"auth-strategies/configs"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Db     DbConfig     `yaml:"db"`
}

type ServerConfig struct {
	Port       int    `yaml:"port"`
	HmacSecret string `yaml:"hmacSecret"`
}

type DbConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func ParseConfig() Config {
	cfg := parseConfigYAML()
	dbHostFromEnv := os.Getenv("POSTGRES_HOST")
	if dbHostFromEnv != "" {
		cfg.Db.Host = dbHostFromEnv
	}
	return cfg
}

// ParseConfig read config.yaml from the working directory and parse it into structs
func parseConfigYAML() Config {
	configFile, err := configs.ConfigYAML.Open("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open embed fs of config.yaml")
	}
	defer configFile.Close()

	cfgBytes, err := io.ReadAll(configFile)
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
