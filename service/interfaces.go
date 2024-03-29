package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
)

//go:generate moq -out ./mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out ./mock/healthcheck.go -pkg mock . HealthChecker

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetKafkaProducer(ctx context.Context, cfg *config.Config, topic string) (kafka.IProducer, error)
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}
