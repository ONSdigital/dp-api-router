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
		So(configuration.BindAddr, ShouldEqual, ":8081")
		So(configuration.HierarchyAPIURL, ShouldEqual, "http://localhost:22600")
		So(configuration.FilterAPIURL, ShouldEqual, "http://localhost:22100")
		So(configuration.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
		So(configuration.RecipeAPIURL, ShouldEqual, "http://localhost:22300")
		So(configuration.ImportAPIURL, ShouldEqual, "http://localhost:21800")
		So(configuration.GracefulShutdown, ShouldEqual, time.Second*5)
	})
}
