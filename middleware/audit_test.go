package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-router/event"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/dp-api-router/schema"
	kafkatest "github.com/ONSdigital/dp-kafka/kafkatest"
	dphttp "github.com/ONSdigital/dp-net/http"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testRequestID          = "myRequest"
	testIdentity           = "myIdentity"
	testCollectionID       = "myCollection"
	testTime               = time.Date(2020, time.April, 26, 7, 5, 52, 123000000, time.UTC)
	testTimeMillis   int64 = 1587884752123
)

func testHandler(statusCode int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(statusCode)
	})
}

func TestAuditHandler(t *testing.T) {

	middleware.Now = func() time.Time {
		return testTime
	}

	Convey("Given a proxied successful request wrapped by an AuditHandler", t, func(c C) {

		// prepare request
		req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		// execute request and wait for audit events
		auditBefore, auditAfter := serveAndCaptureAudit(c, w, req, testHandler(http.StatusOK), true)

		expectedAuditEvent := event.Audit{
			CreatedAt:  testTimeMillis,
			Path:       "/v1/datasets",
			Method:     http.MethodGet,
			QueryParam: "q1=v1&q2=v2",
		}

		Convey("Then status OK is returned", func(c C) {
			c.So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("The expected audit event is sent before proxying the call", func() {
			c.So(auditBefore, ShouldResemble, expectedAuditEvent)
		})

		Convey("And the expected audit event with OK status code is sent after proxying the call", func() {
			expectedAuditEvent.StatusCode = int32(http.StatusOK)
			c.So(auditAfter, ShouldResemble, expectedAuditEvent)
		})
	})

	Convey("Given a proxied successful request with context values for requestID, identity and collectionID, wrapped by an AuditHandler", t, func(c C) {

		// prepare test
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		req = req.WithContext(context.WithValue(req.Context(), dphttp.RequestIdKey, testRequestID))
		req = req.WithContext(context.WithValue(req.Context(), dphttp.CallerIdentityKey, testIdentity))
		req = req.WithContext(context.WithValue(req.Context(), dphttp.CollectionIDHeaderKey, testCollectionID))
		w := httptest.NewRecorder()

		// execute request and wait for audit events
		auditBefore, auditAfter := serveAndCaptureAudit(c, w, req, testHandler(http.StatusOK), true)

		expectedAuditEvent := event.Audit{
			CreatedAt:    testTimeMillis,
			RequestID:    testRequestID,
			Identity:     testIdentity,
			CollectionID: testCollectionID,
			Path:         "/v1/datasets",
			Method:       http.MethodGet,
			QueryParam:   "",
		}

		Convey("Then status OK is returned", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("The expected audit event is sent before proxying the call", func() {
			c.So(auditBefore, ShouldResemble, expectedAuditEvent)
		})

		Convey("And the expected audit event with OK status code is sent after proxying the call", func() {
			expectedAuditEvent.StatusCode = int32(http.StatusOK)
			c.So(auditAfter, ShouldResemble, expectedAuditEvent)
		})
	})

	Convey("Given a proxied NotFound request wrapped by an AuditHandler", t, func(c C) {

		// prepare test
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		// execute request and wait for audit events
		auditBefore, auditAfter := serveAndCaptureAudit(c, w, req, testHandler(http.StatusNotFound), true)

		expectedAuditEvent := event.Audit{
			CreatedAt:  testTimeMillis,
			Path:       "/v1/datasets",
			Method:     http.MethodGet,
			QueryParam: "",
		}

		Convey("Then status NotFound is returned", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("The expected audit event is sent before proxying the call", func() {
			c.So(auditBefore, ShouldResemble, expectedAuditEvent)
		})

		Convey("And the expected audit event with NotFound status code is sent after proxying the call", func() {
			expectedAuditEvent.StatusCode = int32(http.StatusNotFound)
			c.So(auditAfter, ShouldResemble, expectedAuditEvent)
		})
	})

	Convey("Given a proxied call with a failing auditing before the call (inbound)", t, func() {
		// TODO verify 500 and empty body
		So(true, ShouldBeTrue)
	})

	Convey("Given a proxied call with a failing auditing after the call (outbound)", t, func() {
		// TODO verify 500 and empty body
		So(true, ShouldBeTrue)
	})
}

// aux function for testing that serves HTTP, wrapping the provided handler with AuditHandler, and waits for the audit events to be sent
func serveAndCaptureAudit(c C, w http.ResponseWriter, req *http.Request, innerHandler http.Handler, expectAfter bool) (auditBefore, auditAfter event.Audit) {

	// create testing kafka producer and audit handler wrapping the provided handler
	p := kafkatest.NewMessageProducer(true)
	wrapped := middleware.AuditHandler(p)(innerHandler)

	// run HTTP server in a parallel go-routine
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		wrapped.ServeHTTP(w, req)
	}()

	auditBefore = captureAuditEvents(c, p.Channels().Output)
	if expectAfter {
		auditAfter = captureAuditEvents(c, p.Channels().Output)
	}
	wg.Wait()
	return auditBefore, auditAfter
}

func captureAuditEvents(c C, outChan chan []byte) event.Audit {
	messageBytes := <-outChan
	auditEvent := event.Audit{}
	err := schema.AuditEvent.Unmarshal(messageBytes, &auditEvent)
	c.So(err, ShouldBeNil)
	return auditEvent
}
