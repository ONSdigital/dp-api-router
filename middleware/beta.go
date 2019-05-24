package middleware

import (
	"github.com/ONSdigital/go-ns/log"
	"net/http"
	"strings"
)

// BetaApiHandler will return a 404 where enforceBetaRoutes is true and the request is aimed at a non beta domain
func BetaApiHandler(enforceBetaRoutes bool, h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if enforceBetaRoutes && !strings.HasPrefix(r.Host, "api.beta") {

			log.InfoCtx(r.Context(), "beta endpoint requested via a non beta domain, returning 404",
				log.Data{"url": r.URL.String()})

			w.WriteHeader(http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	})
}
