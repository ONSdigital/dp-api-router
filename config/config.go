package config

import (
	"net/http"
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/kelseyhightower/envconfig"
)

// Config contains configurable details for running the service
type Config struct {
	BindAddr                             string         `envconfig:"BIND_ADDR"`
	Version                              string         `envconfig:"VERSION"`
	EnableV1BetaRestriction              bool           `envconfig:"ENABLE_V1_BETA_RESTRICTION"`
	EnablePrivateEndpoints               bool           `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	EnableObservationAPI                 bool           `envconfig:"ENABLE_OBSERVATION_API"`
	EnableAudit                          bool           `envconfig:"ENABLE_AUDIT"`
	EnableZebedeeAudit                   bool           `envconfig:"ENABLE_ZEBEDEE_AUDIT"`
	ZebedeeURL                           string         `envconfig:"ZEBEDEE_URL"`
	HierarchyAPIURL                      string         `envconfig:"HIERARCHY_API_URL"`
	FilterAPIURL                         string         `envconfig:"FILTER_API_URL"`
	FilterFlexAPIURL                     string         `envconfig:"FILTER_FLEX_API_URL"`
	DatasetAPIURL                        string         `envconfig:"DATASET_API_URL"`
	ObservationAPIURL                    string         `envconfig:"OBSERVATION_API_URL"`
	CodelistAPIURL                       string         `envconfig:"CODE_LIST_API_URL"`
	RecipeAPIURL                         string         `envconfig:"RECIPE_API_URL"`
	ImportAPIURL                         string         `envconfig:"IMPORT_API_URL"`
	SearchAPIURL                         string         `envconfig:"SEARCH_API_URL"`
	SearchReindexAPIURL                  string         `envconfig:"SEARCH_REINDEX_API_URL"`
	SearchReindexAPIVersions             []string       `envconfig:"SEARCH_REINDEX_API_VERSIONS"`
	DimensionSearchAPIURL                string         `envconfig:"DIMENSION_SEARCH_API_URL"`
	ImageAPIURL                          string         `envconfig:"IMAGE_API_URL"`
	UploadServiceAPIURL                  string         `envconfig:"UPLOAD_SERVICE_API_URL"`
	DownloadServiceURL                   string         `envconfig:"DOWNLOAD_SERVICE_URL"`
	EnableFilesAPI                       bool           `envconfig:"ENABLE_FILES_API"`
	FilesAPIURL                          string         `envconfig:"FILES_API_URL"`
	IdentityAPIURL                       string         `envconfig:"IDENTITY_API_URL"`
	IdentityAPIVersions                  []string       `envconfig:"IDENTITY_API_VERSIONS"`
	PermissionsAPIURL                    string         `envconfig:"PERMISSIONS_API_URL"`
	PermissionsAPIVersions               []string       `envconfig:"PERMISSIONS_API_VERSIONS"`
	EnvironmentHost                      string         `envconfig:"ENV_HOST"`
	APIPocURL                            string         `envconfig:"API_POC_URL"`
	GracefulShutdown                     time.Duration  `envconfig:"SHUTDOWN_TIMEOUT"`
	AllowedMethods                       []string       `envconfig:"ALLOWED_METHODS"`
	AllowedHeaders                       []string       `envconfig:"ALLOWED_HEADERS"`
	AllowedOrigins                       []string       `envconfig:"ALLOWED_ORIGINS"`
	HealthCheckInterval                  time.Duration  `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout           time.Duration  `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	Brokers                              []string       `envconfig:"KAFKA_ADDR"`
	KafkaVersion                         string         `envconfig:"KAFKA_VERSION"`
	KafkaSecProtocol                     string         `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts                      string         `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert                   string         `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey                    string         `envconfig:"KAFKA_SEC_CLIENT_KEY" json:"-"`
	KafkaSecSkipVerify                   bool           `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	KafkaMaxBytes                        int            `envconfig:"KAFKA_MAX_BYTES"`
	KafkaMinHealthyBrokers               int            `envconfig:"KAFKA_MIN_HEALTHY_BROKERS"`
	AuditTopic                           string         `envconfig:"AUDIT_TOPIC"`
	SessionsAPIURL                       string         `envconfig:"SESSIONS_API_URL"`
	EnableSessionsAPI                    bool           `envconfig:"ENABLE_SESSIONS_API"`
	TopicAPIURL                          string         `envconfig:"TOPIC_API_URL"`
	EnableFeedbackAPI                    bool           `envconfig:"ENABLE_FEEDBACK_API"`
	FeedbackAPIURL                       string         `envconfig:"FEEDBACK_API_URL"`
	FeedbackAPIVersions                  []string       `envconfig:"FEEDBACK_API_VERSIONS"`
	PopulationTypesAPIURL                string         `envconfig:"POPULATION_TYPES_API_URL"`
	EnablePopulationTypesAPI             bool           `envconfig:"ENABLE_POPULATION_TYPES_API"`
	EnableReleaseCalendarAPI             bool           `envconfig:"ENABLE_RELEASE_CALENDAR_API"`
	ReleaseCalendarAPIURL                string         `envconfig:"RELEASE_CALENDAR_API_URL"`
	ReleaseCalendarAPIVersions           []string       `envconfig:"RELEASE_CALENDAR_API_VERSIONS"`
	EnableCantabularMetadataExtractorAPI bool           `envconfig:"ENABLE_CANTABULAR_METADATA_EXTRACTOR_API"`
	CantabularMetadataExtractorAPIURL    string         `envconfig:"CANTABULAR_METADATA_API_URL"`
	ZebedeeClientTimeout                 time.Duration  `envconfig:"ZEBEDEE_CLIENT_TIMEOUT"`
	EnableNLPSearchAPIs                  bool           `envconfig:"ENABLE_NLP_SEARCH_APIS"`
	SearchScrubberAPIURL                 string         `envconfig:"SEARCH_SCRUBBER_API_URL"`
	SearchScrubberAPIVersions            []string       `envconfig:"SEARCH_SCRUBBER_API_VERSIONS"`
	CategoryAPIURL                       string         `envconfig:"CATEGORY_API_URL"`
	CategoryAPIVersions                  []string       `envconfig:"CATEGORY_API_VERSIONS"`
	BerlinAPIURL                         string         `envconfig:"BERLIN_API_URL"`
	BerlinAPIVersions                    []string       `envconfig:"BERLIN_API_VERSIONS"`
	HTTPWriteTimeout                     *time.Duration `envconfig:"HTTP_WRITE_TIMEOUT"`
	OTExporterOTLPEndpoint               string         `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName                        string         `envconfig:"OTEL_SERVICE_NAME"`
	OTBatchTimeout                       time.Duration  `envconfig:"OTEL_BATCH_TIMEOUT"`
	OtelEnabled                          bool           `envconfig:"OTEL_ENABLED"`
	Auth                                 authorisation.Config
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                             ":23200",
		Version:                              "v1",
		EnablePrivateEndpoints:               true,
		EnableV1BetaRestriction:              false,
		EnableObservationAPI:                 false,
		EnableAudit:                          false,
		EnableZebedeeAudit:                   false,
		EnableFilesAPI:                       false,
		ZebedeeURL:                           "http://localhost:8082",
		HierarchyAPIURL:                      "http://localhost:22600",
		FilterAPIURL:                         "http://localhost:22100",
		FilterFlexAPIURL:                     "http://localhost:27100",
		DatasetAPIURL:                        "http://localhost:22000",
		ObservationAPIURL:                    "http://localhost:24500",
		CodelistAPIURL:                       "http://localhost:22400",
		RecipeAPIURL:                         "http://localhost:22300",
		ImportAPIURL:                         "http://localhost:21800",
		SearchAPIURL:                         "http://localhost:23900",
		SearchReindexAPIURL:                  "http://localhost:25700",
		SearchReindexAPIVersions:             []string{"v1"},
		DimensionSearchAPIURL:                "http://localhost:23100",
		DownloadServiceURL:                   "http://localhost:23600",
		ImageAPIURL:                          "http://localhost:24700",
		UploadServiceAPIURL:                  "http://localhost:25100",
		IdentityAPIURL:                       "http://localhost:25600",
		FilesAPIURL:                          "http://localhost:26900",
		IdentityAPIVersions:                  []string{"v1"},
		PermissionsAPIURL:                    "http://localhost:25400",
		PermissionsAPIVersions:               []string{"v1"},
		APIPocURL:                            "http://localhost:3000",
		EnvironmentHost:                      "http://localhost:23200",
		GracefulShutdown:                     5 * time.Second,
		HealthCheckInterval:                  30 * time.Second,
		HealthCheckCriticalTimeout:           90 * time.Second,
		Brokers:                              []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		KafkaVersion:                         "1.0.2",
		AllowedHeaders:                       []string{"Accept", "Accept-Language", "Content-Language", "Origin", "X-Requested-With", "Content-Type", "Authorization"},
		AllowedOrigins:                       []string{"http://localhost:20000", "http://localhost:8081"},
		AllowedMethods:                       []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions},
		KafkaMaxBytes:                        2000000,
		KafkaMinHealthyBrokers:               0,
		AuditTopic:                           "audit",
		SessionsAPIURL:                       "http://localhost:24400",
		EnableSessionsAPI:                    false,
		TopicAPIURL:                          "http://localhost:25300",
		FeedbackAPIURL:                       "http://localhost:28600",
		EnableFeedbackAPI:                    false,
		FeedbackAPIVersions:                  []string{"v1"},
		PopulationTypesAPIURL:                "http://localhost:27300",
		EnablePopulationTypesAPI:             false,
		ReleaseCalendarAPIURL:                "http://localhost:27800",
		EnableReleaseCalendarAPI:             false,
		ReleaseCalendarAPIVersions:           []string{"v1"},
		EnableCantabularMetadataExtractorAPI: false,
		CantabularMetadataExtractorAPIURL:    "http://localhost:28300",
		ZebedeeClientTimeout:                 30 * time.Second,
		EnableNLPSearchAPIs:                  false,
		SearchScrubberAPIURL:                 "http://localhost:28700",
		SearchScrubberAPIVersions:            []string{"v1"},
		CategoryAPIURL:                       "http://localhost:28800",
		CategoryAPIVersions:                  []string{"v1"},
		BerlinAPIURL:                         "http://localhost:28900",
		BerlinAPIVersions:                    []string{"v1"},
		HTTPWriteTimeout:                     nil,
		OTExporterOTLPEndpoint:               "localhost:4317",
		OTServiceName:                        "dp-api-router",
		OTBatchTimeout:                       time.Second * 5,
		OtelEnabled:                          false,
	}

	return cfg, envconfig.Process("", cfg)
}
