package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
)

// BetaApiHandler will return a 404 where enforceBetaRoutes is true and the request is aimed at a non beta domain
func BetaApiHandler(enableBetaRestriction bool, h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if enableBetaRestriction && !isInternalTraffic(r) && !isBetaDomain(r) {

			log.Warn(r.Context(), "beta endpoint requested via a non beta domain, returning 404",
				log.Data{"url": r.URL.String()})

			w.WriteHeader(http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func isBetaDomain(r *http.Request) bool {
	return strings.HasPrefix(r.Host, "api.beta")
}

func isInternalTraffic(r *http.Request) bool {

	// exclude the port from the potential IP address
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		// if we fail to split from the port, just use the original host value
		host = r.Host
	}

	return isValidIP(host) || host == "localhost"
}

func isValidIP(host string) bool {
	return net.ParseIP(host) != nil
}
