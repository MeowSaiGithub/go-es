package config

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"go-es/server"
	"os"

	"github.com/spf13/viper"
	"go-es/internal/esc"
)

// Config is the main configuration struct.
//
// It contains all the configuration for the Go-ES server.
type Config struct {
	// LogLevel is the log level for the server.
	// It can be one of "debug", "info", "warn", "error".
	LogLevel string `mapstructure:"log_level"`
	// DetailError is a boolean that indicates whether the server should return
	// detailed error messages.
	DetailError bool `mapstructure:"detail_error"`
	// Server is the configuration for the server.
	Server *server.Config `mapstructure:"server"`
	// ElasticSearch is the configuration for the Elasticsearch client.
	ElasticSearch *esc.Config `mapstructure:"elastic_search"`
}

// LoadConfig reads the configuration from a given file.
//
// The file path is given by the "config" flag or the environment variable
// "CONFIG_PATH". If neither is given, it defaults to "./config.yaml".
//
// The configuration is validated using the "github.com/go-playground/validator/v10"
// package.
//
// If the configuration contains a CAPath, the CA certificate is read from the
// given file path.
func LoadConfig() (*Config, error) {
	setDefault()
	viper.BindPFlag("config", pflag.Lookup("config"))

	configPath := viper.GetString("config")

	viper.SetConfigFile(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}

	if cfg.ElasticSearch.CAPath != "" {
		cfg.ElasticSearch.CACert, err = loadCACert(cfg.ElasticSearch.CAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read ca file: %w", err)
		}
	}

	return &cfg, nil
}

// loadCACert reads the CA certificate from the given file path.
//
// The file must be in PEM format and contain a valid X.509 certificate.
func loadCACert(path string) ([]byte, error) {
	cert, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read CA file: %w", err)
	}

	block, _ := pem.Decode(cert)
	if block == nil {
		return nil, fmt.Errorf("invalid CA certificate: not in PEM format")
	}

	// Ensure the certificate is valid
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		return nil, fmt.Errorf("invalid CA certificate: %w", err)
	}

	return cert, nil
}

// setDefault sets default values for the configuration.
//
// This function is used to avoid having to write the same default values twice.
func setDefault() {
	viper.SetDefault("config", "config.yaml")
	viper.SetDefault("base_path", "/")
	viper.SetDefault("port", 8080)
	viper.SetDefault("cors_enable", false)
	viper.SetDefault("origins", []string{})
	viper.SetDefault("log_level", "info")
	viper.SetDefault("detail_error", false)
}
