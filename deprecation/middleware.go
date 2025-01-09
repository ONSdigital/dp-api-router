package deprecation

import (
	"fmt"
	"net/http"
	"time"
)

type Deprecation struct {
	Paths   []string
	Date    string
	Link    string
	Message string
	Sunset  string
	Outages []Outage
}

type Outage struct {
	Start time.Time
	End   time.Time
}

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
