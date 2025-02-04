package esc

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
)

// Config is the configuration for the elastic search client.
type Config struct {
	Addresses []string `mapstructure:"addresses" validate:"required"` // addresses is the list of addresses for the elastic search server.
	APIKey    string   `mapstructure:"api_key" validate:"required"`   // apiKey is the api key for the elastic search server.
	CACert    []byte   `mapstructure:"ca_cert"`                       // caCert is the CA certificate used to verify the identity of the elastic search server.
	CAPath    string   `mapstructure:"ca_path"`                       // caPath is the path to the CA certificate.
}

// NewClient returns a new elasticsearch client based on the configuration.
//
// The elastic search client is created with the addresses and api key provided in the configuration.
// The client is then verified by calling the Info() method.
//
// If an error occurs during the creation of the client or the verification of the client, an error
// is returned.
func NewClient(cfg *Config) (*elasticsearch.Client, error) {
	conf := elasticsearch.Config{
		Addresses: cfg.Addresses,
		APIKey:    cfg.APIKey,
	}
	if cfg.CACert != nil {
		conf.CACert = cfg.CACert
	}

	client, err := elasticsearch.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create an es client: %w", err)
	}
	_, err = client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get client info: %w", err)
	}
	return client, err
}
