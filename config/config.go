package config

import "github.com/kelseyhightower/envconfig"

// Config holds all configuration for the application.
type Config struct {
	Port      int    `envconfig:"PORT" default:"8080"`
	LogLevel  string `envconfig:"LOG_LEVEL" default:"info"`
	DBConnStr string `envconfig:"DATABASE_URL" required:"true"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
