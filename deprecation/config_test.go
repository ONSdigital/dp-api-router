package deprecation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//TODO complete unit tests

//// Test config provider that returns a config from a string rather than loading from a file.
//func configFromString(configStr string) func() ([]byte, error) {
//	return func() ([]byte, error) {
//		return []byte(configStr), nil
//	}
//}

func TestParseConfig(t *testing.T) {
	Convey("Given an empty config string", t, func(c C) {
		//TODO implement unit test correctly
		//configString := ""
		//
		//Convey("LoadConfig will return nil with no errors", func() {
		//	deprecations, err := LoadConfig(configFromString(configString))
		//	So(err, ShouldBeNil)
		//	So(deprecations, ShouldBeEmpty)
		//})
	})

	//TODO implement unit test correctly
	//Convey("Given a valid config string", t, func(c C) {
	//	configString := `[{path:}]`
	//})
}
