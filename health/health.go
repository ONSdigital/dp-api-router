package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
)

var hc *healthcheck.HealthCheck

// InitializeHealthCheck initializes the HealthCheck object with startTime now
func InitializeHealthCheck(ctx context.Context, buildTime, gitCommit, version string) {

	versionInfo, err := healthcheck.CreateVersionInfo(
		buildTime,
		gitCommit,
		version,
	)
	if err != nil {
		log.Event(ctx, "failed to obtain versionInfo", log.Error(err))
	}

	hc = &healthcheck.HealthCheck{
		Status:    healthcheck.StatusOK,
		Version:   versionInfo,
		Uptime:    time.Duration(0),
		StartTime: time.Now().UTC(),
		Checks:    []*healthcheck.Check{},
	}
}

// Handler updates the HealthCheck current uptime, marshalls it, and writes it to the ResponseWriter.
func Handler(w http.ResponseWriter, req *http.Request) {

	if hc == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hc.Uptime = time.Since(hc.StartTime) / time.Millisecond

	marshaled, err := json.Marshal(hc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshaled)
}
