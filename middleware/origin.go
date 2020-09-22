package middleware

import (
	"net/http"

	"github.com/ONSdigital/log.go/log"
)

// SetAllowOriginHeader sets the 'Access-Control-Allow-Origin' field for origin domains we allow. Otherwise returning a 401.
func SetAllowOriginHeader(allowedOrigins []string) func(h http.Handler) http.Handler {
	acceptAllOrigins := false
	for _, v := range allowedOrigins {
		if v == "*" {
			acceptAllOrigins = true
			break
		}
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if acceptAllOrigins {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				origin := r.Header.Get("Origin")

				// Only check the origin if it's actually a cross origin request
				if origin != "" {
					acceptedOrigin := ""
					for _, v := range allowedOrigins {

						if v == origin || v == "*" {
							acceptedOrigin = origin
							break
						}
					}

					if acceptedOrigin == "" {
						log.Event(r.Context(), "request received but origin not allowed, returning 401", log.WARN,
							log.Data{"origin": origin, "allowed_origins": allowedOrigins})
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					w.Header().Set("Access-Control-Allow-Origin", acceptedOrigin)
				}
			}

			h.ServeHTTP(w, r)
		})
	}
}
