package service

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/dp-api-router/proxy"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Run initialises the dependencies, proxy router, and starts the http server
func Run(ctx context.Context, buildTime, gitCommit, version string) error {
	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting config", log.FATAL, log.Data{"config": cfg}, log.Error(err))
		return err
	}
	log.Event(ctx, "starting dp-api-router ....", log.INFO, log.Data{"config": cfg})

	if cfg.EnableV1BetaRestriction {
		log.Event(ctx, "beta route restriction is active, /v1 api requests will only be permitted against beta domains", log.INFO)
	}

	// Healthcheck
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "Failed to obtain VersionInfo for healthcheck", log.FATAL, log.Error(err))
		return err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	// Create router and http server
	router := CreateRouter(ctx, cfg, &hc)
	httpServer := server.New(cfg.BindAddr, router)

	// CORS - only allow certain methods in web
	if !cfg.EnablePrivateEndpoints {
		methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
		httpServer.Middleware["CORS"] = handlers.CORS(methodsOk)
	}

	// CORS - only allow specified origins in publishing
	if cfg.EnablePrivateEndpoints {
		httpServer.Middleware["CORS"] = middleware.SetAllowOriginHeader(cfg.AllowedOrigins)
	}

	httpServer.MiddlewareOrder = append(httpServer.MiddlewareOrder, "CORS")
	httpServer.DefaultShutdownTimeout = cfg.GracefulShutdown

	hc.Start(ctx)
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Event(ctx, "failed to close down http server", log.FATAL, log.Data{"config": cfg}, log.Error(err))
		return err
	}
	return nil
}

// CreateRouter creates the router with the required endpoints for proxied APIs
func CreateRouter(ctx context.Context, cfg *config.Config, hc IHealthCheck) *mux.Router {
	router := mux.NewRouter()

	// Healthcheck Endpoint
	router.HandleFunc("/health", hc.Handler)
	router.HandleFunc(fmt.Sprintf("/%s/health", cfg.Version), hc.Handler)

	// Public APIs
	if cfg.EnableObservationAPI {
		observation := proxy.NewAPIProxy(cfg.ObservationAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
		addVersionHandler(router, observation, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations")
	}
	codeList := proxy.NewAPIProxy(cfg.CodelistAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	dataset := proxy.NewAPIProxy(cfg.DatasetAPIURL, cfg.Version, cfg.EnvironmentHost, cfg.ContextURL, cfg.EnableV1BetaRestriction)
	filter := proxy.NewAPIProxy(cfg.FilterAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	hierarchy := proxy.NewAPIProxy(cfg.HierarchyAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	search := proxy.NewAPIProxy(cfg.SearchAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
	addVersionHandler(router, codeList, "/code-lists")
	addVersionHandler(router, dataset, "/datasets")
	addVersionHandler(router, filter, "/filters")
	addVersionHandler(router, filter, "/filter-outputs")
	addVersionHandler(router, hierarchy, "/hierarchies")
	addVersionHandler(router, search, "/search")

	// Private APIs
	if cfg.EnablePrivateEndpoints {
		recipe := proxy.NewAPIProxy(cfg.RecipeAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		importAPI := proxy.NewAPIProxy(cfg.ImportAPIURL, cfg.Version, cfg.EnvironmentHost, "", cfg.EnableV1BetaRestriction)
		addVersionHandler(router, recipe, "/recipes")
		addVersionHandler(router, importAPI, "/jobs")
		addVersionHandler(router, dataset, "/instances")
	}

	// Legacy API
	poc := proxy.NewAPIProxy(cfg.APIPocURL, "", cfg.EnvironmentHost, "", false)
	addLegacyHandler(router, poc, "/ops")
	addLegacyHandler(router, poc, "/dataset")
	addLegacyHandler(router, poc, "/timeseries")
	addLegacyHandler(router, poc, "/search")

	return router
}

func addVersionHandler(router *mux.Router, proxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(fmt.Sprintf("/%s"+path+"{rest:.*}", proxy.Version), proxy.VersionHandle)
}

func addLegacyHandler(router *mux.Router, proxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(path+"{rest:.*}", proxy.VersionHandle)
}
