package middleware

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
)

func TestBetaHandler_WraperHandler(t *testing.T) {

	Convey("beta handler should wrap another handler", t, func() {
		wrapped := BetaApiHandler(true, dummyHandler)
		So(wrapped, ShouldHaveSameTypeAs, dummyHandler)
	})
}

func TestBetaHandler_RestrictionEnabled(t *testing.T) {

	Convey("a request to a beta domain should return 200 status ok", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}

		req.Host = "api.beta"

		w := httptest.NewRecorder()
		wrapped := BetaApiHandler(true, dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)

	})

	Convey("a request to a non beta domain should return 404 status not found", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}

		req.Host = "api.not.beta"

		w := httptest.NewRecorder()
		wrapped := BetaApiHandler(true, dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 404)

	})
}

func TestBetaHandler_RestrictionNotEnabled(t *testing.T) {

	Convey("a request to a non beta domain should return 200 status ok", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}

		req.Host = "api.beta"

		w := httptest.NewRecorder()
		wrapped := BetaApiHandler(false, dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)

	})
}
