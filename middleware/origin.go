package middleware

import "net/http"

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
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", acceptedOrigin)
			h.ServeHTTP(w, r)
		})
	}
}
