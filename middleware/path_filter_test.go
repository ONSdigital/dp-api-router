package middleware_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-router/middleware"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testBodyHc    = []byte{2, 4, 6, 8}
	testHcHandler = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testBodyHc)
		return
	}
)

func TestHealthcheckFilterHandler(t *testing.T) {

	Convey("Given a HealthcheckFilter handler with a healthcheck handler", t, func() {
		// prepare request
		req, err := http.NewRequest(http.MethodGet, "/health", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		// middleware handler under test
		hcFilterHandler := middleware.HealthcheckFilter(testHcHandler)(nil)

		Convey("Then a request with '/health' path results in status OK and healthcheck body, as provided by the healthcheck handler", func(c C) {
			hcFilterHandler.ServeHTTP(w, req)
			c.So(w.Code, ShouldEqual, http.StatusOK)
			b, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			c.So(b, ShouldResemble, testBodyHc)
		})
	})

	Convey("Given a generic handler returning Forbidden status, wrapped by a HealthcheckFilter handler", t, func() {
		// prepare request
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		// middleware handler under test
		hcFilterHandler := middleware.HealthcheckFilter(nil)(testHandler(http.StatusForbidden, testBody))

		Convey("Then a request with path different than '/health' results in status Forbidden and test body, as provided by the generic handler", func(c C) {
			hcFilterHandler.ServeHTTP(w, req)
			c.So(w.Code, ShouldEqual, http.StatusForbidden)
			b, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			c.So(b, ShouldResemble, testBody)
		})
	})

}
