package main

import (
	"fmt"
	"os"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-api-router/proxy"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

func addVersionHandler(router *mux.Router, proxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(fmt.Sprintf("/%s"+path+"{rest:.*}", proxy.Version), proxy.VersionHandle)
}

func addLegacyHandler(router *mux.Router, proxy *proxy.APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(path+"{rest:.*}", proxy.VersionHandle)
}

func main() {
	log.Namespace = "dp-api-router"
	cfg, err := config.Get()
	if err != nil {
		log.Error(err, log.Data{"config": cfg})
		os.Exit(1)
	}
	log.Info("starting dp-api-router ....", log.Data{"config": cfg})
	router := mux.NewRouter()

	// Public APIs
	codeList := proxy.NewAPIProxy(cfg.CodelistAPIURL, cfg.Version)
	dataset := proxy.NewAPIProxy(cfg.DatasetAPIURL, cfg.Version)
	filter := proxy.NewAPIProxy(cfg.FilterAPIURL, cfg.Version)
	hierarchy := proxy.NewAPIProxy(cfg.HierarchyAPIURL, cfg.Version)
	search := proxy.NewAPIProxy(cfg.SearchAPIURL, cfg.Version)
	addVersionHandler(router, codeList, "/code-lists")
	addVersionHandler(router, dataset, "/datasets")
	addVersionHandler(router, filter, "/filters")
	addVersionHandler(router, filter, "/filter-outputs")
	addVersionHandler(router, hierarchy, "/hierarchies")
	addVersionHandler(router, search, "/search")

	// Private APIs
	if cfg.EnablePrivateEndpoints {
		recipe := proxy.NewAPIProxy(cfg.RecipeAPIURL, cfg.Version)
		importAPI := proxy.NewAPIProxy(cfg.ImportAPIURL, cfg.Version)
		addVersionHandler(router, recipe, "/recipes")
		addVersionHandler(router, importAPI, "/jobs")
		addVersionHandler(router, dataset, "/instances")
	}

	// legacy API
	poc := proxy.NewAPIProxy(cfg.APIPocURL, "")
	addLegacyHandler(router, poc, "/ops")
	addLegacyHandler(router, poc, "/dataset")
	addLegacyHandler(router, poc, "/timeseries")
	addLegacyHandler(router, poc, "/search")


	httpServer := server.New(cfg.BindAddr, router)

	// Enable CORS for GET in Web
	if !cfg.EnablePrivateEndpoints {
		methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
		httpServer.Middleware["CORS"] = handlers.CORS(methodsOk)
		httpServer.MiddlewareOrder = append(httpServer.MiddlewareOrder, "CORS")
	}

	httpServer.DefaultShutdownTimeout = cfg.GracefulShutdown
	err = httpServer.ListenAndServe()
	if err != nil {
		log.ErrorC("failed to close down http server", err, log.Data{"config": cfg})
		os.Exit(1)
	}
}
