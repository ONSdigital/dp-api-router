package config

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetReturnsDefaultValues(t *testing.T) {
	t.Parallel()
	Convey("When a loading a configuration, default values are returned", t, func() {
		configuration, err := Get()

		So(err, ShouldBeNil)
		So(configuration, ShouldResemble, &Config{
			BindAddr:                             ":23200",
			Version:                              "v1",
			EnableInterceptor:                    true,
			EnablePrivateEndpoints:               true,
			EnableV1BetaRestriction:              false,
			EnableObservationAPI:                 false,
			EnableAudit:                          false,
			EnableZebedeeAudit:                   false,
			EnableFilesAPI:                       false,
			EnableBundleAPI:                      false,
			ZebedeeURL:                           "http://localhost:8082",
			HierarchyAPIURL:                      "http://localhost:22600",
			FilterAPIURL:                         "http://localhost:22100",
			FilterFlexAPIURL:                     "http://localhost:27100",
			DatasetAPIURL:                        "http://localhost:22000",
			ObservationAPIURL:                    "http://localhost:24500",
			CodelistAPIURL:                       "http://localhost:22400",
			RecipeAPIURL:                         "http://localhost:22300",
			ImportAPIURL:                         "http://localhost:21800",
			ImageAPIURL:                          "http://localhost:24700",
			UploadServiceAPIURL:                  "http://localhost:25100",
			DownloadServiceURL:                   "http://localhost:23600",
			FilesAPIURL:                          "http://localhost:26900",
			IdentityAPIURL:                       "http://localhost:25600",
			BundleAPIURL:                         "http://localhost:29800",
			IdentityAPIVersions:                  []string{"v1"},
			PermissionsAPIURL:                    "http://localhost:25400",
			PermissionsAPIVersions:               []string{"v1"},
			SearchAPIURL:                         "http://localhost:23900",
			DimensionSearchAPIURL:                "http://localhost:23100",
			EnvironmentHost:                      "http://localhost:23200",
			GracefulShutdown:                     5 * time.Second,
			HealthCheckInterval:                  30 * time.Second,
			HealthCheckCriticalTimeout:           90 * time.Second,
			Brokers:                              []string{"localhost:9092", "localhost:9093", "localhost:9094"},
			KafkaVersion:                         "1.0.2",
			KafkaMaxBytes:                        2000000,
			AllowedHeaders:                       []string{"Accept", "Accept-Language", "Content-Language", "Origin", "X-Requested-With", "Content-Type", "Authorization"},
			AllowedOrigins:                       []string{"http://localhost:20000", "http://localhost:8081"},
			AllowedMethods:                       []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions},
			AuditTopic:                           "audit",
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
			RedirectAPIURL:                       "http://localhost:29900",
			HTTPWriteTimeout:                     nil,
			OTExporterOTLPEndpoint:               "localhost:4317",
			OTServiceName:                        "dp-api-router",
			OTBatchTimeout:                       5 * time.Second,
			DeprecationConfigFilePath:            "",
		})
	})
}
