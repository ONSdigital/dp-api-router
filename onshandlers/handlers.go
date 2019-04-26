package onshandlers

import "net/http"

// OriginHandler sets the 'Access-Control-Allow-Origin' field for origin domains we allow. Otherwise returning a 401.
func OriginHandler(allowedOrigins []string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			allowedOrigin := ""
			origin := r.Header.Get("Origin")
			for _, v := range allowedOrigins {
				if v == origin {
					allowedOrigin = origin
				}
			}

			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				h.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		})
	}
}
