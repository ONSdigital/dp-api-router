package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config contains configurable details for running the service
type Config struct {
	BindAddr         string        `envconfig:"BIND_ADDR"`
	HierarchyAPIURL  string        `envconfig:"HIERARCHY_API_URL"`
	FilterAPIURL     string        `envconfig:"FILTER_API_URL"`
	DatasetAPIURL    string        `envconfig:"DATASET_API_URL"`
	CodelistAPIURL   string        `envconfig:"CODELIST_API_URL"`
	RecipeAPIURL     string        `envconfig:"RECIPE_API_URL"`
	ImportAPIURL     string        `envconfig:"IMPORT_API_URL"`
	SearchAPIURL     string        `envconfig:"SEARCH_API_URL"`
	GracefulShutdown time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
}

var configuration *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if configuration == nil {
		configuration = &Config{
			BindAddr:         ":23200",
			HierarchyAPIURL:  "http://localhost:22600",
			FilterAPIURL:     "http://localhost:22100",
			DatasetAPIURL:    "http://localhost:22000",
			CodelistAPIURL:   "http://localhost:22400",
			RecipeAPIURL:     "http://localhost:22300",
			ImportAPIURL:     "http://localhost:21800",
			SearchAPIURL:     "http://localhost:23100",
			GracefulShutdown: 5 * time.Second,
		}
		if err := envconfig.Process("", configuration); err != nil {
			return nil, err
		}
	}
	return configuration, nil
}
