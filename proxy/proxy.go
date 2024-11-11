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

// NewSingleHostReverseProxy is a function that creates a new httputil ReverseProxy, with default transport
var NewSingleHostReverseProxy = func(target *url.URL) IReverseProxy {
	return NewSingleHostReverseProxyWithTransport(target, nil)
}

// NewSingleHostReverseProxyWithTransport is a function that creates a new httputil ReverseProxy, with the supplied transport
var NewSingleHostReverseProxyWithTransport = func(target *url.URL, transport http.RoundTripper) IReverseProxy {
	pxy := httputil.NewSingleHostReverseProxy(target)
	if transport != nil {
		pxy.Transport = transport
	}
	return pxy
}

// Options is a struct that allows optional parameters to be supplied when initialising an API proxy
type Options struct {
	Interceptor bool
}

// NewAPIProxy creates a new APIProxy with a new ReverseProxy for the provided target
func NewAPIProxy(ctx context.Context, target, version, envHost string, enableBetaRestriction bool) *APIProxy {
	return NewAPIProxyWithOptions(ctx, target, version, envHost, enableBetaRestriction, Options{})
}

// NewAPIProxyWithOptions creates a new APIProxy with a new ReverseProxy for the provided target that accepts optional parameters
func NewAPIProxyWithOptions(ctx context.Context, target, version, envHost string, enableBetaRestriction bool, options Options) *APIProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal(ctx, "failed to create url", err, log.Data{"url": target})
		os.Exit(1)
	}

	var transport http.RoundTripper
	if options.Interceptor {
		transport = interceptor.NewRoundTripper(envHost+"/"+version, http.DefaultTransport)
	}

	pxy := NewSingleHostReverseProxyWithTransport(targetURL, transport)
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

	middleware.BetaAPIHandler(p.enableBetaRestriction, p.proxy).ServeHTTP(w, r)
}
