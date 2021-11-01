package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/ONSdigital/dp-api-router/interceptor"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/log.go/v2/log"
)

// APIProxy will forward any requests to an API
type APIProxy struct {
	target                *url.URL
	proxy                 IReverseProxy
	Version               string
	enableBetaRestriction bool
}

// NewSingleHostReverseProxy is a function that creates a new httputil ReverseProxy, and its transport
var NewSingleHostReverseProxy = func(target *url.URL, version, envHost, contextURL string) IReverseProxy {
	pxy := httputil.NewSingleHostReverseProxy(target)
	pxy.Transport = interceptor.NewRoundTripper(envHost+"/"+version, contextURL, http.DefaultTransport)
	return pxy
}

// NewAPIProxy creates a new APIProxy with a new ReverseProxy for the provided target
func NewAPIProxy(ctx context.Context, target, version, envHost, contextURL string, enableBetaRestriction bool) *APIProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal(ctx, "failed to create url", err, log.Data{"url": target})
		os.Exit(1)
	}

	pxy := NewSingleHostReverseProxy(targetURL, version, envHost, contextURL)
	return &APIProxy{
		target:                targetURL,
		proxy:                 pxy,
		Version:               version,
		enableBetaRestriction: enableBetaRestriction}
}

// Handle is a wrapper for proxy ServeHTTP
func (p *APIProxy) Handle(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

// LegacyHandle removes the /v1 path item from the URL and then calls the proxy's ServeHTTP
func (p *APIProxy) LegacyHandle(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "/v1", "", 1)

	middleware.BetaApiHandler(p.enableBetaRestriction, p.proxy).ServeHTTP(w, r)
}
