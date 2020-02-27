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
	"github.com/ONSdigital/log.go/log"
)

// APIProxy will forward any requests to a API
type APIProxy struct {
	target                *url.URL
	proxy                 *httputil.ReverseProxy
	Version               string
	enableBetaRestriction bool
}

func NewAPIProxy(target, version, envHost, contextURL string, enableBetaRestriction bool) *APIProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Event(context.Background(), "failed to create url", log.FATAL, log.Data{"url": target}, log.Error(err))
		os.Exit(1)
	}

	pxy := httputil.NewSingleHostReverseProxy(targetURL)
	pxy.Transport = interceptor.NewRoundTripper(envHost+"/"+version, contextURL, http.DefaultTransport)

	return &APIProxy{
		target:                targetURL,
		proxy:                 pxy,
		Version:               version,
		enableBetaRestriction: enableBetaRestriction}
}

func (p *APIProxy) Handle(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

func (p *APIProxy) VersionHandle(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "/v1", "", 1)

	middleware.BetaApiHandler(p.enableBetaRestriction, p.proxy).ServeHTTP(w, r)
}
