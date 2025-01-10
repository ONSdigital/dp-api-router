package deprecation

import (
	"fmt"
	"net/http"
	"time"
)

// Deprecation is a struct that holds details of an individual deprecation configuration such as the times it is for and
// the paths it applies to. It can optionally contain multiple [Outage]'s
type Deprecation struct {
	Paths   []string
	Date    string
	Link    string
	Message string
	Sunset  string
	Outages []Outage
}

// Outage is a struct covering the start and end times of individual outages
type Outage struct {
	Start time.Time
	End   time.Time
}

// Router is a function that returns a middleware handler which intercepts http traffic and applies the deprecation
// [Middleware] handler to the defined routes in the supplied [Deprecation] configurations. If a route doesn't match it
// is passed through to the underlying handler unmodified.
// If no [Deprecation] configurations are supplied then the underlying handler is returned instead to avoid any
// unnecessary performance overhead.
func Router(deprecations []Deprecation) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if deprecations == nil {
			return next
		}
		mux := http.NewServeMux()
		for _, dep := range deprecations {
			for _, path := range dep.Paths {
				mux.Handle(path, Middleware(dep)(next))
			}
		}
		mux.Handle("/", next)
		return mux
	}
}

// Middleware is a function that returns a middleware handler which intercepts requests and applies headers as per the
// [Deprecation] config. If a configured [Outage] is in force then the handler responds with a
// [http.StatusNotFound] (404) response, otherwise the request is forwarded on to the underlying handler instead.
// Note: this middleware disregards the paths in the [Deprecation] config and applies to all requests it receives. If
// the paths need to be considered, use the [Router] middleware instead of using this middleware directly.
func Middleware(deprecation Deprecation) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			now := time.Now().UTC()

			w.Header().Set("Deprecation", "true")
			if deprecation.Date != "" {
				w.Header().Set("Deprecation", deprecation.Date)
			}

			if deprecation.Link != "" {
				w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"sunset\"", deprecation.Link))
			}

			if deprecation.Sunset != "" {
				w.Header().Set("Sunset", deprecation.Sunset) // Wed, 11 Nov 2020 23:59:59 GMT
			}

			// check if time of request is during a deprecation-outage
			for _, outage := range deprecation.Outages {
				if outage.Start.Before(now) {
					if outage.End.After(now) {
						http.Error(w, deprecation.Message, http.StatusNotFound)
						return
					}
				} else {
					// Outages are sorted by Start time
					break // skip later outages
				}
			}

			h.ServeHTTP(w, req)
		})
	}
}
