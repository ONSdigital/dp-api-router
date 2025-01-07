package deprecation

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type deprecationConfig []struct {
	Paths   []string `json:"paths"`
	Date    string   `json:"date"`
	Sunset  string
	Outages []string
	Link    string
	Msg     string
}

type configProvider func() ([]byte, error)

func ConfigFromFile(filename string) func() ([]byte, error) {
	return func() ([]byte, error) {
		return os.ReadFile(filename)
	}
}

func LoadConfig(cfgProvider configProvider) ([]Deprecation, error) {
	configJSON, err := cfgProvider()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load deprecation config")
	}
	log.Info(context.Background(), "loaded deprecation config", log.Data{"deprecation_json": string(configJSON)})

	var depConfig deprecationConfig
	err = json.Unmarshal(configJSON, &depConfig)
	if err != nil {
		return nil, errors.Wrap(err, "invalid json in deprecation config")
	}
	log.Info(context.Background(), "unmarshalled deprecation config", log.Data{"deprecation_content": depConfig})

	deprecations := make([]Deprecation, len(depConfig))
	for i, config := range depConfig {
		if err := validatePaths(config.Paths...); err != nil {
			return nil, err
		}

		date, err := parseTime(config.Date)
		if err != nil {
			return nil, errors.Wrap(err, "invalid date in deprecation config: "+config.Date)
		}
		datestr := date.Format(time.RFC1123)

		sunset, err := parseTime(config.Sunset)
		if err != nil {
			return nil, errors.Wrap(err, "invalid sunset in deprecation config: "+config.Sunset)
		}
		sunsetstr := sunset.Format(time.RFC1123)

		outages, err := parseOutages(config.Outages)
		if err != nil {
			return nil, errors.Wrap(err, "invalid outages in deprecation config")
		}

		deprecations[i] = Deprecation{
			Paths:   config.Paths,
			Date:    datestr,
			Link:    config.Link,
			Message: config.Msg,
			Sunset:  sunsetstr,
			Outages: outages,
		}
		log.Info(context.Background(), "added deprecation", log.Data{"deprecation": deprecations[i]})
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
			mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {})
		}()
		if err != nil {
			return errors.New(fmt.Sprintf("invalid path spec: '%s' (%v)", path, err))
		}
	}
	return nil
}

func parseTime(timestr string) (time.Time, error) {
	for _, fmt := range []string{time.RFC3339, time.DateOnly, time.DateTime} {
		if time, err := time.Parse(fmt, timestr); err == nil {
			return time, nil
		}
	}
	return time.Time{}, errors.New("invalid time format")
}

func parseOutages(outagestrings []string) ([]Outage, error) {
	outages := make([]Outage, 0, len(outagestrings))
	// convert OutageStrings to Outages
	var err error
	for i, outageStr := range outagestrings {
		if outagePairStr := strings.Split(outageStr, "@"); len(outagePairStr) == 2 {
			var periodLen time.Duration
			if periodLen, err = time.ParseDuration(outagePairStr[0]); err != nil {
				return nil, fmt.Errorf("cannot parse `duration@...` in period %d: %w", i+1, err)
			}

			var periodStart time.Time
			if periodStart, err = parseTime(outagePairStr[1]); err != nil {
				return nil, fmt.Errorf("cannot parse `...@time` in period %d: %w", i+1, err)
			}

			outages = append(outages, Outage{Start: periodStart, End: periodStart.Add(periodLen)})
		} else {
			return nil, fmt.Errorf("expected `duration@time` in period %d", i+1)
		}
	}
	if len(outagestrings) > 1 {
		sort.Slice(outages, func(i, j int) bool {
			return outages[i].Start.Before(outages[j].Start)
		})
	}
	return outages, nil
}
