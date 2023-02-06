package config

import (
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
			ImageAPIURL:                          "http://localhost:24700",
			UploadServiceAPIURL:                  "http://localhost:25100",
			DownloadServiceURL:                   "http://localhost:23600",
			FilesAPIURL:                          "http://localhost:26900",
			IdentityAPIURL:                       "http://localhost:25600",
			IdentityAPIVersions:                  []string{"v1"},
			PermissionsAPIURL:                    "http://localhost:25400",
			PermissionsAPIVersions:               []string{"v1"},
			SearchAPIURL:                         "http://localhost:23900",
			SearchReindexAPIURL:                  "http://localhost:25700",
			SearchReindexAPIVersions:             []string{"v1"},
			DimensionSearchAPIURL:                "http://localhost:23100",
			APIPocURL:                            "http://localhost:3000",
			ContextURL:                           "",
			EnvironmentHost:                      "http://localhost:23200",
			GracefulShutdown:                     5 * time.Second,
			HealthCheckInterval:                  30 * time.Second,
			HealthCheckCriticalTimeout:           90 * time.Second,
			Brokers:                              []string{"localhost:9092", "localhost:9093", "localhost:9094"},
			KafkaVersion:                         "1.0.2",
			KafkaMaxBytes:                        2000000,
			AllowedOrigins:                       []string{"http://localhost:8081"},
			AuditTopic:                           "audit",
			SessionsAPIURL:                       "http://localhost:24400",
			EnableSessionsAPI:                    false,
			TopicAPIURL:                          "http://localhost:25300",
			ArticlesAPIURL:                       "http://localhost:27000",
			ArticlesAPIVersions:                  []string{"v1"},
			FeedbackAPIURL:                       "http://localhost:28600",
			EnableFeedbackAPI:                    false,
			FeedbackAPIVersions:                  []string{"v1"},
			PopulationTypesAPIURL:                "http://localhost:27300",
			EnablePopulationTypesAPI:             false,
			ReleaseCalendarAPIURL:                "http://localhost:27800",
			EnableReleaseCalendarAPI:             false,
			ReleaseCalendarAPIVersions:           []string{"v1"},
			InteractivesAPIURL:                   "http://localhost:27500",
			EnableInteractivesAPI:                false,
			InteractivesAPIVersions:              []string{"v1"},
			EnableMapsAPI:                        false,
			MapsAPIURL:                           "http://localhost:27900",
			MapsAPIVersions:                      []string{"v1"},
			EnableGeodataAPI:                     false,
			GeodataAPIURL:                        "http://localhost:28200",
			GeodataAPIVersions:                   []string{"v1"},
			EnableAreasAPI:                       false,
			AreasAPIURL:                          "http://localhost:25500",
			AreasAPIVersions:                     []string{"v1"},
			EnableCantabularMetadataExtractorAPI: false,
			CantabularMetadataExtractorAPIURL:    "http://localhost:28300",
		})
	})
}
