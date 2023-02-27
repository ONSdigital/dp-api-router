package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
)

func TestBetaHandler(t *testing.T) {
	Convey("beta handler should wrap another handler", t, func() {
		wrapped := BetaAPIHandler(true, dummyHandler)
		So(wrapped, ShouldHaveSameTypeAs, dummyHandler)
	})

	Convey("where beta restrictions are enabled", t, func() {
		Convey("a request to a beta domain should return 200 status ok", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = "api.beta"

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)
		})

		Convey("a request to a non beta domain should return 404 status not found", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = "api.not.beta"

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 404)
		})

		Convey("a request to an internal IP and port should return 200 status", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = "10.201.4.85:80"

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)
		})

		Convey("a request to a non beta host and port should return 404 status", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = "somehost:20100"

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 404)
		})

		Convey("a request to an internal IP should return 200 status", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = "10.201.4.85"

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)
		})

		Convey("a request to localhost should return 200 status", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = localhost

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)
		})

		Convey("a request to localhost with a port should return 200 status", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = "localhost:20300"

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)
		})
	})

	Convey("where beta restrictions are not enabled", t, func() {
		Convey("a request to a non beta domain should return 200 status ok", func() {
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)

			req.Host = "api.not.beta"

			w := httptest.NewRecorder()
			wrapped := BetaAPIHandler(false, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)
		})
	})
}
