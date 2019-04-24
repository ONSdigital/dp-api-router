package config

import (
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/kelseyhightower/envconfig"
)

// Config contains configurable details for running the service
type Config struct {
	BindAddr               string        `envconfig:"BIND_ADDR"`
	Version                string        `envconfig:"VERSION"`
	EnableV1Routes         bool          `envconfig:"ENABLE_V1_ROUTES"`
	EnablePrivateEndpoints bool          `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	HierarchyAPIURL        string        `envconfig:"HIERARCHY_API_URL"`
	FilterAPIURL           string        `envconfig:"FILTER_API_URL"`
	DatasetAPIURL          string        `envconfig:"DATASET_API_URL"`
	CodelistAPIURL         string        `envconfig:"CODE_LIST_API_URL"`
	RecipeAPIURL           string        `envconfig:"RECIPE_API_URL"`
	ImportAPIURL           string        `envconfig:"IMPORT_API_URL"`
	SearchAPIURL           string        `envconfig:"SEARCH_API_URL"`
	ContextURL             string        `envconfig:"CONTEXT_URL"`
	EnvironmentHost        string        `envconfig:"ENV_HOST"`
	APIPocURL              string        `envconfig:"API_POC_URL"`
	GracefulShutdown       time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
}

var configuration *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if configuration == nil {
		configuration = &Config{
			BindAddr:               ":23200",
			Version:                "v1",
			EnablePrivateEndpoints: true,
			EnableV1Routes:         false,
			HierarchyAPIURL:        "http://localhost:22600",
			FilterAPIURL:           "http://localhost:22100",
			DatasetAPIURL:          "http://localhost:22000",
			CodelistAPIURL:         "http://localhost:22400",
			RecipeAPIURL:           "http://localhost:22300",
			ImportAPIURL:           "http://localhost:21800",
			SearchAPIURL:           "http://localhost:23100",
			APIPocURL:              "http://localhost:3000",
			ContextURL:             "",
			EnvironmentHost:        "http://localhost:23200",
			GracefulShutdown:       5 * time.Second,
		}
		if err := envconfig.Process("", configuration); err != nil {
			log.ErrorC("failed to parse configuration", err, log.Data{"config": configuration})
			return nil, err
		}
	}
	return configuration, nil
}
