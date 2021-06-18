package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-api-router/event"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/dp-api-router/proxy"
	"github.com/ONSdigital/dp-api-router/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the API Router
type Service struct {
	Config             *config.Config
	ServiceList        *ExternalServiceList
	KafkaAuditProducer kafka.IProducer
	Server             *dphttp.Server
	HealthCheck        HealthChecker
	ZebedeeClient      *health.Client
}

// Run initialises the dependencies, proxy router, and starts the http server
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (svc *Service, err error) {
	log.Event(ctx, "got service configuration", log.Data{"config": cfg}, log.INFO)

	svc = &Service{
		Config:      cfg,
		ServiceList: serviceList,
	}

	if cfg.EnableV1BetaRestriction {
		log.Event(ctx, "beta route restriction is active, /v1 api requests will only be permitted against beta domains", log.INFO)
	}

	// Create Zebedee client
	svc.ZebedeeClient = health.NewClient("Zebedee", cfg.ZebedeeURL)

	// Get Kafka Audit Producer (only if audit is enabled)
	if cfg.EnableAudit {
		svc.KafkaAuditProducer, err = serviceList.GetKafkaAuditProducer(ctx, cfg)
		if err != nil {
			log.Event(ctx, "could not instantiate kafka audit producer", log.FATAL, log.Error(err))
			return nil, err
		}
	}

	// Healthcheck
	svc.HealthCheck, err = serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return nil, err
	}
	if err := svc.registerCheckers(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	// Create router and http server
	r := CreateRouter(ctx, cfg)
	m := svc.CreateMiddleware(cfg, r)
	svc.Server = dphttp.NewServer(cfg.BindAddr, m.Then(r))

	svc.Server.DefaultShutdownTimeout = cfg.GracefulShutdown
	svc.Server.HandleOSSignals = false

	// kafka error channel logging go-routine
	if cfg.EnableAudit {
		svc.KafkaAuditProducer.Channels().LogErrors(ctx, "kafka Audit Producer")
	}

	// Start healthcheck and run the http server in a new go-routine
	svc.HealthCheck.Start(ctx)
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return svc, nil
}

// CreateMiddleware creates an Alice middleware chain of handlers in the required order
func (svc *Service) CreateMiddleware(cfg *config.Config, router *mux.Router) alice.Chain {

	// Allow health check endpoint to skip any further middleware
	healthCheckFilter := middleware.HealthcheckFilter(svc.HealthCheck.Handler)
	versionedHealthCheckFilter := middleware.VersionedHealthCheckFilter(cfg.Version, svc.HealthCheck.Handler)
	m := alice.New(healthCheckFilter, versionedHealthCheckFilter)

	// Audit - send kafka message to track user requests
	if cfg.EnableAudit {
		auditProducer := event.NewAvroProducer(svc.KafkaAuditProducer.Channels().Output, schema.AuditEvent)
		m = m.Append(middleware.AuditHandler(
			auditProducer,
			svc.ZebedeeClient.Client,
			cfg.ZebedeeURL,
			cfg.Version,
			cfg.EnableZebedeeAudit,
			router))
	}

	if cfg.EnablePrivateEndpoints {
		// CORS - only allow specified origins in publishing
		m = m.Append(middleware.SetAllowOriginHeader(cfg.AllowedOrigins))
	} else {
		// CORS - allow all origin domains, but only allow certain methods in web
		methodsOk := handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete})
		m = m.Append(handlers.CORS(methodsOk))
		m = m.Append(middleware.SetAllowOriginHeader([]string{"*"}))
	}

	return m
}

// CreateRouter creates the router with the required endpoints for proxied APIs
func CreateRouter(ctx context.Context, cfg *config.Config) *mux.Router {
	router := mux.NewRouter()

	// Public APIs
	if cfg.EnableObservationAPI {
		observation := proxy.NewAPIProxy(cfg.ObservationAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
		addTransitionalHandler(router, observation, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations")
	}
	if cfg.EnableTopicAPI {
		topic := proxy.NewAPIProxy(cfg.TopicAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
		addTransitionalHandler(router, topic, "/topics")
	}
	codeList := proxy.NewAPIProxy(cfg.CodelistAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	dataset := proxy.NewAPIProxy(cfg.DatasetAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
	filter := proxy.NewAPIProxy(cfg.FilterAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	hierarchy := proxy.NewAPIProxy(cfg.HierarchyAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	search := proxy.NewAPIProxy(cfg.SearchAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	dimensionSearch := proxy.NewAPIProxy(cfg.DimensionSearchAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	image := proxy.NewAPIProxy(cfg.ImageAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	addTransitionalHandler(router, codeList, "/code-lists")
	addTransitionalHandler(router, dataset, "/datasets")
	addTransitionalHandler(router, filter, "/filters")
	addTransitionalHandler(router, filter, "/filter-outputs")
	addTransitionalHandler(router, hierarchy, "/hierarchies")
	addTransitionalHandler(router, search, "/search")
	addTransitionalHandler(router, dimensionSearch, "/dimension-search")
	addTransitionalHandler(router, image, "/images")

	// Private APIs
	fmt.Println("private endpoint config - "+fmt.Sprintf("%v", cfg.EnablePrivateEndpoints))
	if cfg.EnablePrivateEndpoints {
		recipe := proxy.NewAPIProxy(cfg.RecipeAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		importAPI := proxy.NewAPIProxy(cfg.ImportAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		uploadServiceAPI := proxy.NewAPIProxy(cfg.UploadServiceAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		identityAPI := proxy.NewAPIProxy(cfg.IdentityAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addTransitionalHandler(router, recipe, "/recipes")
		addTransitionalHandler(router, importAPI, "/jobs")
		addTransitionalHandler(router, dataset, "/instances")
		addTransitionalHandler(router, uploadServiceAPI, "/upload")
		for _, version := range cfg.IdentityAPIVersions {
			addVersionedHandler(router, identityAPI, version, "/tokens")
			addVersionedHandler(router, identityAPI, version, "/users")
			addVersionedHandler(router, identityAPI, version, "/groups")
			addVersionedHandler(router, identityAPI, version, "/password-reset")
		}

		// Feature flag for Sessions API
		if cfg.EnableSessionsAPI {
			session := proxy.NewAPIProxy(cfg.SessionsAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
			addTransitionalHandler(router, session, "/sessions")
		}
	}

	// Legacy API
	poc := proxy.NewAPIProxy(cfg.APIPocURL, "", cfg.EnvironmentHost, "", false)
	addLegacyHandler(router, poc, "/ops")
	addLegacyHandler(router, poc, "/dataset")
	addLegacyHandler(router, poc, "/timeseries")
	addLegacyHandler(router, poc, "/search")

	zebedee := proxy.NewAPIProxy(cfg.ZebedeeURL, cfg.Version, cfg.EnvironmentHost, "", false)
	router.NotFoundHandler = http.HandlerFunc(zebedee.LegacyHandle)

	return router
}

func addVersionedHandler(router *mux.Router, proxy *proxy.APIProxy, version string, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc("/"+version+path+"{rest:.*}", proxy.Handle)
}

func addTransitionalHandler(router *mux.Router, proxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(fmt.Sprintf("/%s"+path+"{rest:.*}", proxy.Version), proxy.LegacyHandle)
}

func addLegacyHandler(router *mux.Router, proxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(path+"{rest:.*}", proxy.LegacyHandle)
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdown
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.Server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}

		// Close Kafka Audit Producer, if present
		if svc.ServiceList.KafkaAuditProducer {
			if err := svc.KafkaAuditProducer.Close(ctx); err != nil {
				log.Event(ctx, "failed to stop kafka audit producer", log.Error(err), log.ERROR)
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "shutdown timed out", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds all the necessary checkers to healthcheck. Please, only call this function after all dependencies are instanciated
func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	if err = svc.HealthCheck.AddCheck("Zebedee", svc.ZebedeeClient.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add zebedee checker", log.ERROR, log.Error(err))
	}

	if svc.Config.EnableAudit {
		if err = svc.HealthCheck.AddCheck("Kafka Audit Producer", svc.KafkaAuditProducer.Checker); err != nil {
			hasErrors = true
			log.Event(ctx, "failed to add kafka audit producer checker", log.ERROR, log.Error(err))
		}
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
