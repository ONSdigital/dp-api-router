package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

var hc *healthcheck.HealthCheck

// ErrNotInitialized error returned when trying to use HealthCheck without initializing it.
var ErrNotInitialized = errors.New("Healthcheck objct was not initialized")

// InitializeHealthCheck initializes the HealhCheck object with startTime now
func InitializeHealthCheck() *healthcheck.HealthCheck {
	hc = &healthcheck.HealthCheck{
		Status:    healthcheck.StatusOK,
		Version:   "aaa",
		Uptime:    time.Duration(0),
		StartTime: time.Now().UTC(),
		Checks:    []healthcheck.Check{},
	}
	return hc
}

// Handler updates the HealthCheck current uptime, marshalls it, and writes it to the ResponseWriter.
func Handler(w http.ResponseWriter, req *http.Request) {

	if hc == nil {
		// TODO log ErrNotInitialized
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hc.Uptime = time.Since(hc.StartTime) / time.Millisecond

	marshaled, err := json.Marshal(hc)
	if err != nil {
		// TODO log Marhalling error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshaled)
}
