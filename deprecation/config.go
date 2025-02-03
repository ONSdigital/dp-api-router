package deprecation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type deprecationConfig []struct {
	Paths   []string `json:"paths"`
	Date    string   `json:"date"`
	Sunset  string   `json:"sunset"`
	Outages []string `json:"outages"`
	Link    string   `json:"link"`
	Msg     string   `json:"msg"`
}

type loaderFunction func() ([]byte, error)

// LoadConfig is a function that triggers the load of a deprecation configuration and parses its content. It takes in a
// function  that returns the loaded bytes (eg. a function that loads content from disk) and returns a slice of
// [Deprecation] structs as per the contents of the loaded bytes.
func LoadConfig(loader loaderFunction) ([]Deprecation, error) {
	configJSON, err := loader()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load deprecation config")
	}

	if len(configJSON) == 0 {
		return nil, nil
	}

	var depConfig deprecationConfig
	err = json.Unmarshal(configJSON, &depConfig)
	if err != nil {
		return nil, errors.Wrap(err, "invalid json in deprecation config")
	}

	deprecations := make([]Deprecation, len(depConfig))
	for i, config := range depConfig {
		if err := validatePaths(config.Paths...); err != nil {
			return nil, err
		}

		date, err := parseTime(config.Date)
		if err != nil {
			return nil, errors.Wrap(err, "invalid date in deprecation config: "+config.Date)
		}
		dateUnix := fmt.Sprintf("@%d", date.Unix())

		sunset, err := parseTime(config.Sunset)
		if err != nil {
			return nil, errors.Wrap(err, "invalid sunset in deprecation config: "+config.Sunset)
		}
		sunsetstr := sunset.Format(time.RFC1123)

		if date.After(sunset) {
			return nil, errors.New("deprecation date must not be later than sunset")
		}

		outages, err := parseOutages(config.Outages)
		if err != nil {
			return nil, errors.Wrap(err, "invalid outages in deprecation config")
		}

		deprecations[i] = Deprecation{
			Paths:    config.Paths,
			DateUnix: dateUnix,
			Link:     config.Link,
			Message:  config.Msg,
			Sunset:   sunsetstr,
			Outages:  outages,
		}
	}

	return deprecations, nil
}

// http.ServeMux panics if it receives an invalid path so we'll do a dry-run here to test for invalid paths and trap
// the panic instead
func validatePaths(paths ...string) error {
	for _, path := range paths {
		var err any
		func() {
			defer func() { err = recover() }()
			mux := http.NewServeMux()
			mux.HandleFunc(path, func(_ http.ResponseWriter, _ *http.Request) {})
		}()
		if err != nil {
			return fmt.Errorf("invalid path spec: '%s' (%v)", path, err)
		}
	}
	return nil
}

func parseTime(timeStr string) (time.Time, error) {
	for _, timeFmt := range []string{time.RFC3339, time.DateOnly, time.DateTime} {
		if parsedTime, err := time.Parse(timeFmt, timeStr); err == nil {
			return parsedTime, nil
		}
	}
	return time.Time{}, errors.New("invalid time format")
}

func parseOutages(outagestrings []string) ([]Outage, error) {
	outages := make([]Outage, 0, len(outagestrings))
	// convert OutageStrings to Outages
	var err error
	for i, outageStr := range outagestrings {
		outagePairStr := strings.Split(outageStr, "@")

		var duration, timeStr string
		switch len(outagePairStr) {
		case 1:
			timeStr = outagePairStr[0]
		case 2:
			duration = outagePairStr[0]
			timeStr = outagePairStr[1]
		default:
			return nil, fmt.Errorf("invalid outage, expected `[duration@]time` in period %d", i+1)
		}

		var periodStart time.Time
		if periodStart, err = parseTime(timeStr); err != nil {
			return nil, fmt.Errorf("cannot parse `...@time` in period %d: %w", i+1, err)
		}

		if duration == "" {
			outages = append(outages, Outage{Start: periodStart, End: nil})
		} else {
			periodLen, err := time.ParseDuration(duration)
			if err != nil {
				return nil, fmt.Errorf("cannot parse `duration@...` in period %d: %w", i+1, err)
			}
			periodEnd := periodStart.Add(periodLen)
			outages = append(outages, Outage{Start: periodStart, End: &periodEnd})
		}
	}

	if len(outagestrings) > 1 {
		sort.Slice(outages, func(i, j int) bool {
			return outages[i].Start.Before(outages[j].Start)
		})
	}
	return outages, nil
}
