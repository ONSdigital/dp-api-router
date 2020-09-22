package middleware

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
)

func TestBetaHandler(t *testing.T) {

	Convey("beta handler should wrap another handler", t, func() {
		wrapped := BetaApiHandler(true, dummyHandler)
		So(wrapped, ShouldHaveSameTypeAs, dummyHandler)
	})

	Convey("where beta restrictions are enabled", t, func() {

		Convey("a request to a beta domain should return 200 status ok", func() {
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)

			req.Host = "api.beta"

			w := httptest.NewRecorder()
			wrapped := BetaApiHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)

		})

		Convey("a request to a non beta domain should return 404 status not found", func() {
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)

			req.Host = "api.not.beta"

			w := httptest.NewRecorder()
			wrapped := BetaApiHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 404)

		})

		Convey("a request to an internal IP and port should return 200 status", func() {
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)

			req.Host = "10.201.4.85:80"

			w := httptest.NewRecorder()
			wrapped := BetaApiHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)

		})

		Convey("a request to a non beta host and port should return 404 status", func() {
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)

			req.Host = "somehost:20100"

			w := httptest.NewRecorder()
			wrapped := BetaApiHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 404)

		})

		Convey("a request to an internal IP should return 200 status", func() {
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)

			req.Host = "10.201.4.85"

			w := httptest.NewRecorder()
			wrapped := BetaApiHandler(true, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)

		})
	})

	Convey("where beta restrictions are not enabled", t, func() {

		Convey("a request to a non beta domain should return 200 status ok", func() {
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)

			req.Host = "api.not.beta"

			w := httptest.NewRecorder()
			wrapped := BetaApiHandler(false, dummyHandler)

			wrapped.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)

		})
	})
}
