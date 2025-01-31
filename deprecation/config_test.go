package deprecation

import (
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// Test loader that returns a config from a string rather than loading from a file.
func loaderFromString(configStr string) func() ([]byte, error) {
	return func() ([]byte, error) {
		return []byte(configStr), nil
	}
}

func TestParseConfig(t *testing.T) {
	Convey("Given an empty config string", t, func() {
		configString := ""

		Convey("LoadConfig will return nil with no errors", func() {
			deprecations, err := LoadConfig(loaderFromString(configString))
			So(err, ShouldBeNil)
			So(deprecations, ShouldBeEmpty)
		})
	})

	Convey("Given a config that fails to load", t, func() {
		loaderError := errors.New("loader error")
		errLoader := func() ([]byte, error) {
			return nil, loaderError
		}

		Convey("LoadConfig will return nil with loading error", func() {
			deprecations, err := LoadConfig(errLoader)
			So(err, ShouldNotBeNil)
			So(deprecations, ShouldBeEmpty)
			So(err.Error(), ShouldEqual, "unable to load deprecation config: "+loaderError.Error())
		})
	})

	Convey("Given invalid json as config", t, func() {
		configString := "{not valid json at all!!"

		Convey("LoadConfig will return nil with parsing err", func() {
			deprecations, err := LoadConfig(loaderFromString(configString))
			So(err, ShouldNotBeNil)
			So(deprecations, ShouldBeEmpty)
			So(err.Error(), ShouldEqual, "invalid json in deprecation config: invalid character 'n' looking for beginning of object key string")
		})
	})

	Convey("Given a range of config test cases", t, func() {
		type testCase struct {
			name      string
			json      string
			wanted    []Deprecation
			wantedErr string
		}
		tcs := []testCase{
			{
				name:      "With valid JSON but not an array as config",
				json:      "{}",
				wanted:    nil,
				wantedErr: "invalid json in deprecation config: json: cannot unmarshal object into Go value of type deprecation.deprecationConfig",
			},
			{
				name:      "With valid but empty json",
				json:      "[]",
				wanted:    []Deprecation{},
				wantedErr: "",
			},
			{
				name: "With a single deprecation",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/",
                             "/timeseries/"
                           ],
                           "date": "2024-09-10T10:00:00Z",
                           "sunset": "2024-10-14",
                           "outages": [
                             "24h@2024-12-31",
                             "12h@2025-01-01T10:00:00Z"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         }
                       ]`,
				wanted: []Deprecation{
					{
						Paths:    []string{"/ops/", "/dataset/", "/timeseries/"},
						DateUnix: "@1725962400", // Tue, 10 Sep 2024 10:00:00 UTC
						Link:     "https://developer.ons.gov.uk/retirement/v0api/",
						Message:  "Some test message 123 !@£",
						Sunset:   "Mon, 14 Oct 2024 00:00:00 UTC",
						Outages: []Outage{
							{
								Start: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
								End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
							},
							{
								Start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
								End:   time.Date(2025, 1, 1, 22, 0, 0, 0, time.UTC),
							},
						},
					},
				},
				wantedErr: "",
			},
			{
				name: "With two deprecations",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/",
                             "/timeseries/"
                           ],
                           "date": "2024-09-10T10:00:00Z",
                           "sunset": "2024-10-14",
                           "outages": [
                             "24h@2024-12-31",
                             "12h@2025-01-01T10:00:00Z"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         },
                         {
                           "paths": [
                             "/another/path"
                           ],
                           "date": "2023-01-02T10:47:20Z",
                           "sunset": "2024-02-29",
                           "outages": [
                             "24h@2028-02-28 13:00:00",
                             "17m@2023-02-28T13:01:02Z",
                             "24s@2024-02-28"

                           ],
                           "link": "https://ons.gov.uk/",
                           "msg": "!Some other test message 123"
                         }
                       ]`,
				wanted: []Deprecation{
					{
						Paths:    []string{"/ops/", "/dataset/", "/timeseries/"},
						DateUnix: "@1725962400", // Tue, 10 Sep 2024 10:00:00 UTC
						Link:     "https://developer.ons.gov.uk/retirement/v0api/",
						Message:  "Some test message 123 !@£",
						Sunset:   "Mon, 14 Oct 2024 00:00:00 UTC",
						Outages: []Outage{
							{
								Start: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
								End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
							},
							{
								Start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
								End:   time.Date(2025, 1, 1, 22, 0, 0, 0, time.UTC),
							},
						},
					},
					{
						Paths:    []string{"/another/path"},
						DateUnix: "@1672656440", // Mon, 02 Jan 2023 10:47:20 UTC
						Link:     "https://ons.gov.uk/",
						Message:  "!Some other test message 123",
						Sunset:   "Thu, 29 Feb 2024 00:00:00 UTC",
						Outages: []Outage{
							{
								Start: time.Date(2023, 2, 28, 13, 1, 2, 0, time.UTC),
								End:   time.Date(2023, 2, 28, 13, 18, 2, 0, time.UTC),
							},
							{
								Start: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC),
								End:   time.Date(2024, 2, 28, 0, 0, 24, 0, time.UTC),
							},
							{
								Start: time.Date(2028, 2, 28, 13, 0, 0, 0, time.UTC),
								End:   time.Date(2028, 2, 29, 13, 0, 0, 0, time.UTC),
							},
						},
					},
				},
				wantedErr: "",
			},
			{
				name: "With invalid path",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/",
                             "totally invalid"
                           ],
                           "date": "2024-09-10T10:00:00Z",
                           "sunset": "2024-10-14",
                           "outages": [
                             "24h@2024-12-31",
                             "12h@2025-01-01T10:00:00Z"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         }
                       ]`,
				wanted:    nil,
				wantedErr: `invalid path spec: 'totally invalid' (parsing "totally invalid": at offset 8: host/path missing /)`,
			},
			{
				name: "With invalid date",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/"
                           ],
                           "date": "this is not a date",
                           "sunset": "2024-10-14",
                           "outages": [
                             "24h@2024-12-31",
                             "12h@2025-01-01T10:00:00Z"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         }
                       ]`,
				wanted:    nil,
				wantedErr: `invalid date in deprecation config: this is not a date: invalid time format`,
			},
			{
				name: "With invalid sunset",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/"
                           ],
                           "date": "2025-01-01T10:00:00Z",
                           "sunset": "absolute nonsense",
                           "outages": [
                             "24h@2024-12-31",
                             "12h@2025-01-01T10:00:00Z"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         }
                       ]`,
				wanted:    nil,
				wantedErr: `invalid sunset in deprecation config: absolute nonsense: invalid time format`,
			},
			{
				name: "With invalid outage duration",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/"
                           ],
                           "date": "2025-01-01T10:00:00Z",
                           "sunset": "2025-01-09 10:00:00",
                           "outages": [
                             "24z@2024-12-31",
                             "blah@2025-01-01T10:00:00Z"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         }
                       ]`,
				wanted:    nil,
				wantedErr: "invalid outages in deprecation config: cannot parse `duration@...` in period 1: time: unknown unit \"z\" in duration \"24z\"",
			},
			{
				name: "With invalid outage time",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/"
                           ],
                           "date": "2025-01-01T10:00:00Z",
                           "sunset": "2025-01-09 10:00:00",
                           "outages": [
                             "24h@2024-12-31",
                             "17m@nonsense date"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         }
                       ]`,
				wanted:    nil,
				wantedErr: "invalid outages in deprecation config: cannot parse `...@time` in period 2: invalid time format",
			},
			{
				name: "With invalid outage instance",
				json: `[
                         {
                           "paths": [
                             "/ops/",
                             "/dataset/"
                           ],
                           "date": "2025-01-01T10:00:00Z",
                           "sunset": "2025-01-09 10:00:00",
                           "outages": [
                             "12h 2025-01-01"
                           ],
                           "link": "https://developer.ons.gov.uk/retirement/v0api/",
                           "msg": "Some test message 123 !@£"
                         }
                       ]`,
				wanted:    nil,
				wantedErr: "invalid outages in deprecation config: invalid outage, expected `duration@time` in period 1",
			},
		}

		for _, tc := range tcs {
			Convey(tc.name, func() {
				deprecations, err := LoadConfig(loaderFromString(tc.json))
				if tc.wantedErr == "" {
					So(err, ShouldBeNil)
				} else {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, tc.wantedErr)
				}
				So(deprecations, ShouldResemble, tc.wanted)
			})
		}
	})
}
