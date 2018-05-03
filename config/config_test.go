package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetRetrunsDefaultValues(t *testing.T) {
	t.Parallel()
	Convey("When a loading a configuration, default values are return", t, func() {
		configuration, err := Get()
		So(err, ShouldBeNil)
		So(configuration.BindAddr, ShouldEqual, ":23200")
		So(configuration.Version, ShouldEqual, "v1")
		So(configuration.EnablePrivateEndpoints, ShouldEqual, true)
		So(configuration.HierarchyAPIURL, ShouldEqual, "http://localhost:22600")
		So(configuration.FilterAPIURL, ShouldEqual, "http://localhost:22100")
		So(configuration.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
		So(configuration.RecipeAPIURL, ShouldEqual, "http://localhost:22300")
		So(configuration.ImportAPIURL, ShouldEqual, "http://localhost:21800")
		So(configuration.SearchAPIURL, ShouldEqual, "http://localhost:23100")
		So(configuration.APIPocURL, ShouldEqual, "http://localhost:3000")
		So(configuration.EnvironmentHost, ShouldEqual, "http://localhost:23200")
		So(configuration.GracefulShutdown, ShouldEqual, time.Second*5)
	})
}
