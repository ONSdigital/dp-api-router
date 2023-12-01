package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-api-router/event"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/dp-api-router/proxy"
	"github.com/ONSdigital/dp-api-router/schema"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/log.go/v2/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
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
	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

	svc = &Service{
		Config:      cfg,
		ServiceList: serviceList,
	}

	if cfg.EnableV1BetaRestriction {
		log.Info(ctx, "beta route restriction is active, /v1 api requests will only be permitted against beta domains")
	}

	// Create Zebedee client
	svc.ZebedeeClient = health.NewClientWithClienter("Zebedee", cfg.ZebedeeURL, dphttp.ClientWithTimeout(dphttp.NewClient(), cfg.ZebedeeClientTimeout))

	// Get Kafka Audit Producer (only if audit is enabled)
	if cfg.EnableAudit {
		svc.KafkaAuditProducer, err = serviceList.GetKafkaAuditProducer(ctx, cfg)
		if err != nil {
			log.Fatal(ctx, "could not instantiate kafka audit producer", err)
			return nil, err
		}
	}

	// Healthcheck
	svc.HealthCheck, err = serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}
	if err := svc.registerCheckers(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	// Create router and http server
	r := CreateRouter(ctx, cfg)
	m := svc.CreateMiddleware(cfg, r)
	r.Use(otelmux.Middleware(cfg.OTServiceName))
	svc.Server = dphttp.NewServer(cfg.BindAddr, m.Then(r))

	svc.Server.DefaultShutdownTimeout = cfg.GracefulShutdown
	svc.Server.HandleOSSignals = false

	// kafka error channel logging go-routine
	if cfg.EnableAudit {
		svc.KafkaAuditProducer.LogErrors(ctx)
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
			router,
			cfg.Auth,
		))
	}

	methodsOk := handlers.AllowedMethods(cfg.AllowedMethods)
	headersOk := handlers.AllowedHeaders(cfg.AllowedHeaders)
	originsOk := handlers.AllowedOrigins(cfg.AllowedOrigins)

	m = m.Append(handlers.CORS(originsOk, headersOk, methodsOk))

	return m
}

// CreateRouter creates the router with the required endpoints for proxied APIs
// The preferred approach for new APIs is to use `addVersionedHandlers` and include the version on downstream API routes
func CreateRouter(ctx context.Context, cfg *config.Config) *mux.Router {
	router := mux.NewRouter()

	// Public APIs
	if cfg.EnableObservationAPI {
		observation := proxy.NewAPIProxy(ctx, cfg.ObservationAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
		addTransitionalHandler(router, observation, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations")
	}

	topic := proxy.NewAPIProxy(ctx, cfg.TopicAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
	addTransitionalHandler(router, topic, "/topics")
	addTransitionalHandler(router, topic, "/navigation")

	codeList := proxy.NewAPIProxy(ctx, cfg.CodelistAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	dataset := proxy.NewAPIProxy(ctx, cfg.DatasetAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
	filter := proxy.NewAPIProxy(ctx, cfg.FilterAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	filterFlex := proxy.NewAPIProxy(ctx, cfg.FilterFlexAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	hierarchy := proxy.NewAPIProxy(ctx, cfg.HierarchyAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	search := proxy.NewAPIProxy(ctx, cfg.SearchAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	dimensionSearch := proxy.NewAPIProxy(ctx, cfg.DimensionSearchAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	image := proxy.NewAPIProxy(ctx, cfg.ImageAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)

	if cfg.EnableAreasAPI {
		areas := proxy.NewAPIProxy(ctx, cfg.AreasAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addVersionedHandlers(router, areas, cfg.AreasAPIVersions, "/areas")
	}
	if cfg.EnableReleaseCalendarAPI {
		releaseCalendar := proxy.NewAPIProxy(ctx, cfg.ReleaseCalendarAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addVersionedHandlers(router, releaseCalendar, cfg.ReleaseCalendarAPIVersions, "/releases")
	}
	if cfg.EnableFeedbackAPI {
		feedback := proxy.NewAPIProxy(ctx, cfg.FeedbackAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addVersionedHandlers(router, feedback, cfg.FeedbackAPIVersions, "/feedback")
	}

	addTransitionalHandler(router, filterFlex, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/json")
	addTransitionalHandler(router, filterFlex, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/census-observations")
	addTransitionalHandler(router, filterFlex, "/custom/filters")

	addTransitionalHandler(router, codeList, "/code-lists")
	addTransitionalHandler(router, dataset, "/datasets")
	addTransitionalHandler(router, filter, "/filters")
	addTransitionalHandler(router, filter, "/filter-outputs")
	addTransitionalHandler(router, hierarchy, "/hierarchies")
	addTransitionalHandler(router, search, "/search")
	addTransitionalHandler(router, dimensionSearch, "/dimension-search")
	addTransitionalHandler(router, image, "/images")

	if cfg.EnablePopulationTypesAPI {
		populationTypesAPI := proxy.NewAPIProxy(ctx, cfg.PopulationTypesAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addTransitionalHandler(router, populationTypesAPI, "/population-types")
	}
	if cfg.EnableMapsAPI {
		mapsProxy := proxy.NewAPIProxy(ctx, cfg.MapsAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addVersionedHandlers(router, mapsProxy, cfg.MapsAPIVersions, "/maps")
	}
	if cfg.EnableGeodataAPI {
		geodataAPIproxy := proxy.NewAPIProxy(ctx, cfg.GeodataAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addVersionedHandlers(router, geodataAPIproxy, cfg.GeodataAPIVersions, "/geodata")
	}

	if cfg.EnableFilesAPI {
		downloadService := proxy.NewAPIProxy(ctx, cfg.DownloadServiceURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		filesAPI := proxy.NewAPIProxy(ctx, cfg.FilesAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)

		addTransitionalHandler(router, filesAPI, "/files")
		addTransitionalHandler(router, downloadService, "/downloads-new")
	}

	if cfg.EnableNLPSearchAPIs {
		searchScrubberAPIProxy := proxy.NewAPIProxy(ctx, cfg.SearchScrubberAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		categoryAPIProxy := proxy.NewAPIProxy(ctx, cfg.CategoryAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		berlinAPIProxy := proxy.NewAPIProxy(ctx, cfg.BerlinAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)

		addVersionedHandlers(router, searchScrubberAPIProxy, cfg.SearchScrubberAPIVersions, "/scrubber")
		addVersionedHandlers(router, categoryAPIProxy, cfg.CategoryAPIVersions, "/categories")
		addVersionedHandlers(router, berlinAPIProxy, cfg.BerlinAPIVersions, "/berlin")
	}

	// Private APIs
	if cfg.EnablePrivateEndpoints {
		recipe := proxy.NewAPIProxy(ctx, cfg.RecipeAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		importAPI := proxy.NewAPIProxy(ctx, cfg.ImportAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		uploadServiceAPI := proxy.NewAPIProxy(ctx, cfg.UploadServiceAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		identityAPI := proxy.NewAPIProxy(ctx, cfg.IdentityAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		permissionsAPIProxy := proxy.NewAPIProxy(ctx, cfg.PermissionsAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		searchReindexAPI := proxy.NewAPIProxy(ctx, cfg.SearchReindexAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addTransitionalHandler(router, recipe, "/recipes")
		addTransitionalHandler(router, importAPI, "/jobs")
		addTransitionalHandler(router, dataset, "/instances")
		addTransitionalHandler(router, uploadServiceAPI, "/upload")
		addVersionedHandlers(router, identityAPI, cfg.IdentityAPIVersions, "/tokens")
		addVersionedHandlers(router, identityAPI, cfg.IdentityAPIVersions, "/users")
		addVersionedHandlers(router, identityAPI, cfg.IdentityAPIVersions, "/groups")
		addVersionedHandlers(router, identityAPI, cfg.IdentityAPIVersions, "/password-reset")
		addVersionedHandlers(router, permissionsAPIProxy, cfg.PermissionsAPIVersions, "/policies")
		addVersionedHandlers(router, permissionsAPIProxy, cfg.PermissionsAPIVersions, "/roles")
		addVersionedHandlers(router, permissionsAPIProxy, cfg.PermissionsAPIVersions, "/permissions-bundle")
		addVersionedHandlers(router, searchReindexAPI, cfg.SearchReindexAPIVersions, "/search-reindex-jobs")

		// Feature flag for Sessions API
		if cfg.EnableSessionsAPI {
			session := proxy.NewAPIProxy(ctx, cfg.SessionsAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
			addTransitionalHandler(router, session, "/sessions")
		}

		// Feature flag for Files API
		if cfg.EnableFilesAPI {
			addTransitionalHandler(router, uploadServiceAPI, "/upload-new")
		}

		// Feature flag for Cantabular Metadata Extractor API
		if cfg.EnableCantabularMetadataExtractorAPI {
			cantMetadataExtractorAPIProxy := proxy.NewAPIProxy(ctx, cfg.CantabularMetadataExtractorAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
			addTransitionalHandler(router, cantMetadataExtractorAPIProxy, "/cantabular-metadata")
		}
	}

	// Legacy API
	poc := proxy.NewAPIProxy(ctx, cfg.APIPocURL, "", cfg.EnvironmentHost, "", false)
	addLegacyHandler(router, poc, "/ops")
	addLegacyHandler(router, poc, "/dataset")
	addLegacyHandler(router, poc, "/timeseries")
	addLegacyHandler(router, poc, "/search")

	zebedee := proxy.NewAPIProxy(ctx, cfg.ZebedeeURL, cfg.Version, cfg.EnvironmentHost, "", false)
	router.NotFoundHandler = http.HandlerFunc(zebedee.LegacyHandle)

	return router
}

func addVersionedHandlers(router *mux.Router, apiProxy *proxy.APIProxy, versions []string, path string) {
	// Proxy any request after the path given to the target address
	for _, version := range versions {
		router.HandleFunc("/"+version+path+"{rest:.*}", apiProxy.Handle)
	}
}

func addTransitionalHandler(router *mux.Router, apiProxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(fmt.Sprintf("/%s"+path+"{rest:$|/.*}", apiProxy.Version), apiProxy.LegacyHandle)
}

func addLegacyHandler(router *mux.Router, apiProxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(path+"{rest:.*}", apiProxy.LegacyHandle)
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdown
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
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
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// Close Kafka Audit Producer, if present
		if svc.ServiceList.KafkaAuditProducer {
			if err := svc.KafkaAuditProducer.Close(ctx); err != nil {
				log.Error(ctx, "failed to stop kafka audit producer", err)
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

// registerCheckers adds all the necessary checkers to healthcheck. Please, only call this function after all dependencies are instanciated
func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	if err = svc.HealthCheck.AddCheck("Zebedee", svc.ZebedeeClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add zebedee checker", err)
	}

	if svc.Config.EnableAudit {
		if err = svc.HealthCheck.AddCheck("Kafka Audit Producer", svc.KafkaAuditProducer.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "failed to add kafka audit producer checker", err)
		}
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
