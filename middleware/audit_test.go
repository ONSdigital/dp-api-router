package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	dphttp "github.com/ONSdigital/dp-net/http"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	testRequestID    = "myRequest"
	testIdentity     = "myIdentity"
	testCollectionID = "myCollection"
)

func TestAuditHandler(t *testing.T) {

	Convey("Given a proxied successful request wrapped by an AuditHandler", t, func() {
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		wrapped := AuditHandler(dummyHandler)
		wrapped.ServeHTTP(w, req)

		Convey("Then status OK is returned", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected Audit events are sent to kafka before and after the call", func() {
			// TODO validate that expected event is sent, once implemented
		})
	})

	Convey("Given a proxied successful request with context values for requestID, identity and collectionID, wrapped by an AuditHandler", t, func() {
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		req = req.WithContext(context.WithValue(req.Context(), dphttp.RequestIdKey, testRequestID))
		req = req.WithContext(context.WithValue(req.Context(), dphttp.CallerIdentityKey, testIdentity))
		req = req.WithContext(context.WithValue(req.Context(), dphttp.CollectionIDHeaderKey, testCollectionID))
		w := httptest.NewRecorder()
		wrapped := AuditHandler(dummyHandler)
		wrapped.ServeHTTP(w, req)

		Convey("Then status OK is returned", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("And the expected Audit events are sent to kafka before and after the call", func() {
			// TODO validate that expected event is sent, once implemented
		})
	})

	Convey("Given a proxied NotFound request wrapped by an AuditHandler", t, func() {
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		wrapped := AuditHandler(NotFoundHandler)
		wrapped.ServeHTTP(w, req)

		Convey("Then status NotFound is returned", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("And the expected Audit events are sent to kafka before and after the call", func() {
			// TODO validate that expected event is sent, once implemented
		})
	})
}
