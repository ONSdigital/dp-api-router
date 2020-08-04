package middleware

import (
	"net/http"
)

// Allowed provides a list of methods for which the handler should be executed
type Allowed struct {
	Methods []string
	Handler func(w http.ResponseWriter, req *http.Request)
}

// isMethodAllowed determines if a method is allowed or not
func (a *Allowed) isMethodAllowed(method string) bool {
	for _, s := range a.Methods {
		if method == s {
			return true
		}
	}
	return false
}

// HealthcheckFilter is a middleware that executed the health endpoint directly (handler provided as a parameter),
// skipping any further middleware handlers
var HealthcheckFilter = func(hcHandler func(w http.ResponseWriter, req *http.Request)) func(h http.Handler) http.Handler {
	return PathFilter(map[string]Allowed{
		"/health": {
			Methods: []string{http.MethodGet},
			Handler: hcHandler,
		},
	})
}

// PathFilter is a middleware that executes allowed endpoints, skipping any further middleware handler
func PathFilter(allowedMap map[string]Allowed) func(h http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			if allowed, ok := allowedMap[req.URL.Path]; ok && allowed.isMethodAllowed(req.Method) {
				allowed.Handler(w, req)
				return
			}

			nextHandler.ServeHTTP(w, req)
		})
	}
}
