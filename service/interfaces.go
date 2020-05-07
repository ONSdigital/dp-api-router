package service

import (
	"context"
	"net/http"
)

//go:generate moq -out ./mock/healthcheck.go -pkg mock . IHealthCheck

// IHealthCheck represens the required methods from HealthCheck
type IHealthCheck interface {
	Start(ctx context.Context)
	Stop()
	Handler(w http.ResponseWriter, req *http.Request)
}
