package middleware_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-router/event"
	eventmock "github.com/ONSdigital/dp-api-router/event/mock"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/dp-api-router/schema"

	"github.com/ONSdigital/dp-kafka/kafkatest"
	dphttp "github.com/ONSdigital/dp-net/http"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testRequestID          = "myRequest"
	testIdentity           = "myIdentity"
	testCollectionID       = "myCollection"
	testTime               = time.Date(2020, time.April, 26, 7, 5, 52, 123000000, time.UTC)
	testTimeMillis   int64 = 1587884752123
	testBody               = []byte{1, 2, 3, 4}
	errMarshal             = errors.New("avro marshal error")
	errCopy                = errors.New("io.Copy error")
)

func testHandler(statusCode int, body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(body)
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

		// audit handler under test
		p := kafkatest.NewMessageProducer(true)
		auditProducer := event.NewAvroProducer(p.Channels().Output, schema.AuditEvent)
		auditHandler := middleware.AuditHandler(auditProducer)(testHandler(http.StatusOK, testBody))

		// execute request and wait for audit events
		auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

		expectedAuditEvent := event.Audit{
			CreatedAt:  testTimeMillis,
			Path:       "/v1/datasets",
			Method:     http.MethodGet,
			QueryParam: "q1=v1&q2=v2",
		}

		Convey("Then status OK and expected body is returned", func(c C) {
			c.So(w.Code, ShouldEqual, http.StatusOK)
			b, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			c.So(b, ShouldResemble, testBody)
		})

		Convey("The expected audit events are sent before and after proxying the call", func() {
			c.So(auditEvents[0], ShouldResemble, expectedAuditEvent)
			expectedAuditEvent.StatusCode = int32(http.StatusOK)
			c.So(auditEvents[1], ShouldResemble, expectedAuditEvent)
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

		// audit handler under test
		p := kafkatest.NewMessageProducer(true)
		auditProducer := event.NewAvroProducer(p.Channels().Output, schema.AuditEvent)
		auditHandler := middleware.AuditHandler(auditProducer)(testHandler(http.StatusOK, testBody))

		// execute request and wait for audit events
		auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

		expectedAuditEvent := event.Audit{
			CreatedAt:    testTimeMillis,
			RequestID:    testRequestID,
			Identity:     testIdentity,
			CollectionID: testCollectionID,
			Path:         "/v1/datasets",
			Method:       http.MethodGet,
			QueryParam:   "",
		}

		Convey("Then status OK and the expected body is returned", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
			b, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			c.So(b, ShouldResemble, testBody)
		})

		Convey("The expected audit event is sent before and after proxying the call", func() {
			c.So(auditEvents[0], ShouldResemble, expectedAuditEvent)
			expectedAuditEvent.StatusCode = int32(http.StatusOK)
			c.So(auditEvents[1], ShouldResemble, expectedAuditEvent)
		})
	})

	Convey("Given a proxied NotFound request wrapped by an AuditHandler", t, func(c C) {

		// prepare test
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		// audit handler under test
		p := kafkatest.NewMessageProducer(true)
		auditProducer := event.NewAvroProducer(p.Channels().Output, schema.AuditEvent)
		auditHandler := middleware.AuditHandler(auditProducer)(testHandler(http.StatusNotFound, []byte{}))

		// execute request and wait for audit events
		auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

		expectedAuditEvent := event.Audit{
			CreatedAt:  testTimeMillis,
			Path:       "/v1/datasets",
			Method:     http.MethodGet,
			QueryParam: "",
		}

		Convey("Then status NotFound and an empty body is returned", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
			b, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			c.So(b, ShouldResemble, []byte{})
		})

		Convey("The expected audit event is sent before and after proxying the call", func() {
			c.So(auditEvents[0], ShouldResemble, expectedAuditEvent)
			expectedAuditEvent.StatusCode = int32(http.StatusNotFound)
			c.So(auditEvents[1], ShouldResemble, expectedAuditEvent)
		})
	})
}

func TestAuditHandlerFailure(t *testing.T) {

	Convey("Given a proxied call with a failing auditing before the call (inbound)", t, func(c C) {

		// prepare test
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		failingMarshaller := &eventmock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return []byte{}, errMarshal
			},
		}

		// audit handler under test
		p := kafkatest.NewMessageProducer(true)
		auditProducer := event.NewAvroProducer(p.Channels().Output, failingMarshaller)
		auditHandler := middleware.AuditHandler(auditProducer)(nil)

		// execute request and wait for audit events
		serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

		Convey("Then status NotFound and empty body is returned", func(c C) {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			b, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			c.So(b, ShouldResemble, []byte{})
		})
	})

	Convey("Given a proxied call with a failing auditing after the call (outbound)", t, func(c C) {

		// prepare test
		req, err := http.NewRequest("GET", "/v1/datasets", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		nMarshalCall := 0
		failingMarshaller := &eventmock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				if nMarshalCall == 0 {
					nMarshalCall++
					auditEvent := event.Audit{
						CreatedAt: testTimeMillis,
						Path:      "/v1/datasets",
						Method:    http.MethodGet,
					}
					b, err := schema.AuditEvent.Marshal(auditEvent)
					c.So(err, ShouldBeNil)
					return b, nil
				}
				return []byte{}, errMarshal
			},
		}

		// audit handler under test
		p := kafkatest.NewMessageProducer(true)
		auditProducer := event.NewAvroProducer(p.Channels().Output, failingMarshaller)
		auditHandler := middleware.AuditHandler(auditProducer)(testHandler(http.StatusOK, testBody))

		// execute request and wait for audit events
		serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 1)

		Convey("Then status NotFound and empty body is returned", func(c C) {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			b, err := ioutil.ReadAll(w.Body)
			So(err, ShouldBeNil)
			c.So(b, ShouldResemble, []byte{})
		})
	})
}

// aux function for testing that serves HTTP, wrapping the provided handler with AuditHandler,
// and waits for the number of expected audit events, which are then returned in an array
func serveAndCaptureAudit(c C, w http.ResponseWriter, req *http.Request, auditHandler http.Handler, outChan chan []byte, numExpectedMessages int) (auditEvents []event.Audit) {

	// run HTTP server in a parallel go-routine
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		auditHandler.ServeHTTP(w, req)
	}()

	// capture audit events from kafka output channel
	auditEvents = []event.Audit{}
	for i := 0; i < numExpectedMessages; i++ {
		auditEvents = append(auditEvents, captureAuditEvent(c, outChan))
	}

	wg.Wait()
	return auditEvents
}

// captureAuditEvent reads the provided channel and unmarshals the bytes to an auditEvent
func captureAuditEvent(c C, outChan chan []byte) event.Audit {
	messageBytes := <-outChan
	auditEvent := event.Audit{}
	err := schema.AuditEvent.Unmarshal(messageBytes, &auditEvent)
	c.So(err, ShouldBeNil)
	return auditEvent
}
