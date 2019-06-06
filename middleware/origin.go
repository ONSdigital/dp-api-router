package middleware

import (
	"github.com/ONSdigital/go-ns/log"
	"net/http"
)

// SetAllowOriginHeader sets the 'Access-Control-Allow-Origin' field for origin domains we allow. Otherwise returning a 401.
func SetAllowOriginHeader(allowedOrigins []string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			acceptedOrigin := ""
			origin := r.Header.Get("Origin")
			for _, v := range allowedOrigins {
				if v == origin {
					acceptedOrigin = origin
					break
				}
			}

			if acceptedOrigin == "" {
				log.InfoCtx(r.Context(), "request received but origin not allowed, returning 401",
					log.Data{"origin": origin, "allowed_origins": allowedOrigins})
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", acceptedOrigin)
			h.ServeHTTP(w, r)
		})
	}
}
