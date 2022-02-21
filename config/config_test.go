package config

import (
	"github.com/hashicorp/go-uuid"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetRetrunsDefaultValues(t *testing.T) {
	t.Parallel()
	Convey("When a loading a configuration, default values are return", t, func() {
		configuration, err := Get()
		So(err, ShouldBeNil)
		So(configuration, ShouldResemble, &Config{
			BindAddr:                   ":23200",
			Version:                    "v1",
			EnablePrivateEndpoints:     true,
			EnableV1BetaRestriction:    false,
			EnableObservationAPI:       false,
			EnableAudit:                false,
			EnableZebedeeAudit:         false,
			ZebedeeURL:                 "http://localhost:8082",
			HierarchyAPIURL:            "http://localhost:22600",
			FilterAPIURL:               "http://localhost:22100",
			DatasetAPIURL:              "http://localhost:22000",
			ObservationAPIURL:          "http://localhost:24500",
			CodelistAPIURL:             "http://localhost:22400",
			RecipeAPIURL:               "http://localhost:22300",
			ImportAPIURL:               "http://localhost:21800",
			ImageAPIURL:                "http://localhost:24700",
			UploadServiceAPIURL:        "http://localhost:25100",
			IdentityAPIURL:             "http://localhost:25600",
			IdentityAPIVersions:        []string{"v1"},
			SearchAPIURL:               "http://localhost:23900",
			DimensionSearchAPIURL:      "http://localhost:23100",
			APIPocURL:                  "http://localhost:3000",
			ContextURL:                 "",
			EnvironmentHost:            "http://localhost:23200",
			GracefulShutdown:           5 * time.Second,
			HealthCheckInterval:        30 * time.Second,
			HealthCheckCriticalTimeout: 90 * time.Second,
			Brokers:                    []string{"localhost:9092", "localhost:9093", "localhost:9094"},
			KafkaVersion:               "1.0.2",
			KafkaMaxBytes:              2000000,
			AllowedOrigins:             []string{"http://localhost:8081"},
			AuditTopic:                 "audit",
			SessionsAPIURL:             "http://localhost:24400",
			EnableSessionsAPI:          false,
			TopicAPIURL:                "http://localhost:25300",
			EnableTopicAPI:             false,
			ArticlesAPIURL:             "http://localhost:27000",
			EnableArticlesAPI:          false,
			PopulationTypesAPIURL:      "http://localhost:27300",
			EnablePopulationTypesAPI:   false,
			ReleaseCalendarAPIURL:      "http://localhost:27800",
			EnableReleaseCalendarAPI:   false,
		})
	})
}

func TestEnvironmentVariableBinding(t *testing.T) {

	Convey("PopulationTypesAPIURL should be bound to POPULATION_TYPES_API_URL", t, func() {
		expectedValue, _ := uuid.GenerateUUID()
		undo := setEnvVar("POPULATION_TYPES_API_URL", expectedValue)
		defer undo()
		configuration, err := Get()
		So(err, ShouldBeNil)
		So(configuration.PopulationTypesAPIURL, ShouldEqual, expectedValue)
	})

	Convey("EnablePopulationTypesAPI should be bound to ENABLE_POPULATION_TYPES_API", t, func() {
		undo := setEnvVar("ENABLE_POPULATION_TYPES_API", "true")
		defer undo()
		configuration, err := Get()
		So(err, ShouldBeNil)
		So(configuration.EnablePopulationTypesAPI, ShouldBeTrue)
	})
}

func setEnvVar(name string, value string) func() {
	existingValue, isValueSet := os.LookupEnv(name)
	setenvOrPanic(name, value)
	return func() {
		if isValueSet {
			setenvOrPanic(name, existingValue)
		} else {
			unsetenvOrPanic(name)
		}
		// remove the cached configuration so that it can be re-read on next Get
		Flush()
	}
}

func unsetenvOrPanic(name string) {
	if err := os.Unsetenv(name); err != nil {
		panic(err)
	}
}

func setenvOrPanic(name string, value string) {
	if err := os.Setenv(name, value); err != nil {
		panic(err)
	}
}
