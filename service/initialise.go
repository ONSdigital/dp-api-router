package service

import (
	"context"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck        bool
	KafkaAuditProducer bool
	Init               Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Init: initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHealthCheck creates a healthcheck with versionInfo and sets the HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// GetKafkaAuditProducer returns a kafka producer for auditing
func (e *ExternalServiceList) GetKafkaAuditProducer(ctx context.Context, cfg *config.Config) (kafkaProducer kafka.IProducer, err error) {
	kafkaProducer, err = e.Init.DoGetKafkaProducer(ctx, cfg, cfg.AuditTopic)
	if err != nil {
		return nil, err
	}
	e.KafkaAuditProducer = true
	return kafkaProducer, nil
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// DoGetKafkaProducer creates a kafka producer for the provided broker addresses, topic and envMax values in config
func (e *Init) DoGetKafkaProducer(ctx context.Context, cfg *config.Config, topic string) (kafka.IProducer, error) {
	pConfig := &kafka.ProducerConfig{
		KafkaVersion:    &cfg.KafkaVersion,
		MaxMessageBytes: &cfg.KafkaMaxBytes,
		Topic:           topic,
		BrokerAddrs:     cfg.Brokers,
	}
	if cfg.KafkaSecProtocol == "TLS" {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(cfg.KafkaSecCACerts, cfg.KafkaSecClientCert, cfg.KafkaSecClientKey, cfg.KafkaSecSkipVerify)
	}
	return kafka.NewProducer(ctx, pConfig)
}

// DoGetAuthorisationMiddleware creates authorisation middleware for the given config
func (e *Init) DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return authorisation.NewFeatureFlaggedMiddleware(ctx, authorisationConfig, authorisationConfig.JWTVerificationPublicKeys)
}
