package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

// APIProxy will forward any requests to a API
type APIProxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func newAPIProxy(target string) *APIProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.ErrorC("failed to create url", err, log.Data{"url": target})
		os.Exit(1)
	}
	return &APIProxy{target: targetURL, proxy: httputil.NewSingleHostReverseProxy(targetURL)}
}

func (p *APIProxy) handle(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

func addHandler(router *mux.Router, proxy *APIProxy, path string) {
	// Proxy any request after the path given to the target address
	router.HandleFunc(path+"{rest:.*}", proxy.handle)
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
	codeList := newAPIProxy(cfg.CodelistAPIURL)
	dataset := newAPIProxy(cfg.DatasetAPIURL)
	filter := newAPIProxy(cfg.FilterAPIURL)
	hierarchy := newAPIProxy(cfg.HierarchyAPIURL)
	addHandler(router, codeList, "/code-lists")
	addHandler(router, dataset, "/datasets")
	addHandler(router, filter, "/filters")
	addHandler(router, hierarchy, "/hierarchies")

	// Private APIs
	recipe := newAPIProxy(cfg.RecipeAPIURL)
	importAPI := newAPIProxy(cfg.ImportAPIURL)
	addHandler(router, recipe, "/recipes")
	addHandler(router, importAPI, "/jobs")

	httpServer := server.New(cfg.BindAddr, router)
	httpServer.DefaultShutdownTimeout = cfg.GracefulShutdown
	err = httpServer.ListenAndServe()
	if err != nil {
		log.ErrorC("failed to close down http server", err, log.Data{"config": cfg})
		os.Exit(1)
	}
}
