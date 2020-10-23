package config

import (
	"context"
	"time"

	"github.com/ONSdigital/log.go/log"
	"github.com/kelseyhightower/envconfig"
)

// Config contains configurable details for running the service
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Version                    string        `envconfig:"VERSION"`
	EnableV1BetaRestriction    bool          `envconfig:"ENABLE_V1_BETA_RESTRICTION"`
	EnablePrivateEndpoints     bool          `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	EnableObservationAPI       bool          `envconfig:"ENABLE_OBSERVATION_API"`
	EnableAudit                bool          `envconfig:"ENABLE_AUDIT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	HierarchyAPIURL            string        `envconfig:"HIERARCHY_API_URL"`
	FilterAPIURL               string        `envconfig:"FILTER_API_URL"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	ObservationAPIURL          string        `envconfig:"OBSERVATION_API_URL"`
	CodelistAPIURL             string        `envconfig:"CODE_LIST_API_URL"`
	RecipeAPIURL               string        `envconfig:"RECIPE_API_URL"`
	ImportAPIURL               string        `envconfig:"IMPORT_API_URL"`
	SearchAPIURL               string        `envconfig:"SEARCH_API_URL"`
	ImageAPIURL                string        `envconfig:"IMAGE_API_URL"`
	UploadServiceAPIURL        string        `envconfig:"UPLOAD_SERVICE_API_URL"`
	ContextURL                 string        `envconfig:"CONTEXT_URL"`
	EnvironmentHost            string        `envconfig:"ENV_HOST"`
	APIPocURL                  string        `envconfig:"API_POC_URL"`
	GracefulShutdown           time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
	AllowedOrigins             []string      `envconfig:"ALLOWED_ORIGINS"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"`
	KafkaMaxBytes              int           `envconfig:"KAFKA_MAX_BYTES"`
	AuditTopic                 string        `envconfig:"AUDIT_TOPIC"`
	SessionsAPIURL             string        `envconfig:"SESSIONS_API_URL"`
	EnableSessionsAPI          bool          `envconfig:"ENABLE_SESSIONS_API"`
	TopicAPIURL                string        `envconfig:"TOPIC_API_URL"`
	EnableTopicAPI             bool          `envconfig:"ENABLE_TOPIC_API"`
}

var configuration *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if configuration == nil {
		configuration = &Config{
			BindAddr:                   ":23200",
			Version:                    "v1",
			EnablePrivateEndpoints:     true,
			EnableV1BetaRestriction:    false,
			EnableObservationAPI:       false,
			EnableAudit:                false,
			ZebedeeURL:                 "http://localhost:8082",
			HierarchyAPIURL:            "http://localhost:22600",
			FilterAPIURL:               "http://localhost:22100",
			DatasetAPIURL:              "http://localhost:22000",
			ObservationAPIURL:          "http://localhost:24500",
			CodelistAPIURL:             "http://localhost:22400",
			RecipeAPIURL:               "http://localhost:22300",
			ImportAPIURL:               "http://localhost:21800",
			SearchAPIURL:               "http://localhost:23100",
			ImageAPIURL:                "http://localhost:24700",
			UploadServiceAPIURL:        "http://localhost:25100",
			APIPocURL:                  "http://localhost:3000",
			ContextURL:                 "",
			EnvironmentHost:            "http://localhost:23200",
			GracefulShutdown:           5 * time.Second,
			HealthCheckInterval:        30 * time.Second,
			HealthCheckCriticalTimeout: 90 * time.Second,
			Brokers:                    []string{"localhost:9092"},
			AllowedOrigins:             []string{"http://localhost:8081"},
			KafkaMaxBytes:              2000000,
			AuditTopic:                 "audit",
			SessionsAPIURL:             "http://localhost:24400",
			EnableSessionsAPI:          false,
			EnableTopicAPI:             true,
		}
		if err := envconfig.Process("", configuration); err != nil {
			log.Event(context.Background(), "failed to parse configuration", log.ERROR, log.Data{"config": configuration}, log.Error(err))
			return nil, err
		}
	}
	return configuration, nil
}
