package deprecation

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseConfig(t *testing.T) {
	Convey("Given an empty config string", t, func(c C) {
		configString := ""

		Convey("LoadConfig will return nl with no errors", func() {
			deprecations, err := LoadConfig(configString)
			So(err, ShouldBeNil)
			So(deprecations, ShouldBeEmpty)
		})
	})

	Convey("Given a valid config string", t, func(c C) {
		configString := `[{path:}]`
	})
}
