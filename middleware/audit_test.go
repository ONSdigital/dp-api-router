package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-router/middleware/mock"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-api-router/event"
	eventmock "github.com/ONSdigital/dp-api-router/event/mock"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/dp-api-router/schema"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testVersionPrefix            = "v1"
	testZebedeeURL               = "myZebedeeURL"
	testRequestID                = "myRequest"
	testIdentity                 = "myIdentity"
	testCollectionID             = "myCollection"
	testFlorenceToken            = "myFlorenceToken"
	testServiceAuthToken         = "myServiceAuthToken"
	testTimeInbound              = time.Date(2020, time.April, 26, 7, 5, 52, 123000000, time.UTC)
	testTimeMillisInbound  int64 = 1587884752123
	testTimeOutbound             = time.Date(2020, time.April, 26, 7, 5, 52, 456000000, time.UTC)
	testTimeMillisOutbound int64 = 1587884752456
	testBody                     = []byte{1, 2, 3, 4}
	errMarshal                   = errors.New("avro marshal error")
)

// valid identity response for testing
var testIdentityResponse = &dprequest.IdentityResponse{
	Identifier: testIdentity,
}

// utility function to generate handlers for testing, which return the provided status code and body
func testHandler(statusCode int, body []byte, c C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(statusCode)
		_, err := w.Write(body)
		c.So(err, ShouldBeNil)
	})
}

// utility function to create a producer and valid audit handler
func createValidAuditHandler() (kafka.IProducer, func(h http.Handler) http.Handler) {
	cliMock := createHTTPClientMock(http.StatusOK, testIdentityResponse)
	p := kafkatest.NewMessageProducer(true)
	auditProducer := event.NewAvroProducer(p.Channels().Output, schema.AuditEvent)
	enableZebedeeAudit := true
	return p, middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, nil)
}

// utility function to create a producer and an audit handler that fails to marshal and send events
func createFailingAuditHandler() (kafka.IProducer, func(h http.Handler) http.Handler) {
	cliMock := createHTTPClientMock(http.StatusOK, testIdentityResponse)
	failingMarshaller := &eventmock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return []byte{}, errMarshal
		},
	}
	p := kafkatest.NewMessageProducer(true)
	auditProducer := event.NewAvroProducer(p.Channels().Output, failingMarshaller)
	enableZebedeeAudit := true
	return p, middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, nil)
}

// utility function to generate Clienter mocks
func createHTTPClientMock(retCode int, retBody interface{}) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{}
		},
		SetPathsWithNoRetriesFunc: func([]string) {
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			body, _ := json.Marshal(retBody)
			return &http.Response{
				StatusCode: retCode,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
			}, nil
		},
	}
}

func TestGenerateAuditEvent(t *testing.T) {

	Convey("Given a mocked time.Now", t, func(c C) {

		middleware.Now = func() time.Time {
			return testTimeInbound
		}

		Convey("A request with query paramters generates a valid audit event, with the expected values", func() {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
			So(err, ShouldBeNil)
			e := middleware.GenerateAuditEvent(req)
			So(*e, ShouldResemble, event.Audit{
				CreatedAt:  testTimeMillisInbound,
				Path:       "/v1/datasets",
				Method:     http.MethodGet,
				QueryParam: "q1=v1&q2=v2",
			})
		})

		Convey("A request with query paramters including escaped characters generates a valid audit event, with the expected unescaped values", func() {
			req, err := http.NewRequest(http.MethodGet, "/v1/data?lang=en\u0026uri=%2Fhealth", nil)
			So(err, ShouldBeNil)
			e := middleware.GenerateAuditEvent(req)
			So(*e, ShouldResemble, event.Audit{
				CreatedAt:  testTimeMillisInbound,
				Path:       "/v1/data",
				Method:     http.MethodGet,
				QueryParam: "lang=en&uri=/health",
			})
		})

		Convey("A request with query paramters including incorrectly escaped characters defaults to the raw query value when generating the audit event", func() {
			req, err := http.NewRequest(http.MethodGet, "/v1/data?uri=%wxhealth", nil)
			So(err, ShouldBeNil)
			e := middleware.GenerateAuditEvent(req)
			So(*e, ShouldResemble, event.Audit{
				CreatedAt:  testTimeMillisInbound,
				Path:       "/v1/data",
				Method:     http.MethodGet,
				QueryParam: "uri=%wxhealth",
			})
		})

	})
}

func TestAuditHandlerHeaders(t *testing.T) {

	Convey("Given deterministic inbound and outbound timestamps", t, func(c C) {

		isInbound := true
		middleware.Now = func() time.Time {
			if isInbound {
				isInbound = false
				return testTimeInbound
			}
			return testTimeOutbound
		}

		Convey("An incoming request with no auth headers", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			Convey("And a valid audit handler without downstream", func(c C) {
				p, a := createValidAuditHandler()
				auditHandler := a(nil)

				// execute request and wait for 1 audit event
				auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 1)

				Convey("Then status Unauthorised and empty body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusUnauthorized)
					b, err := ioutil.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, []byte{})
				})

				Convey("Then the expected inbound audit event is sent", func() {
					c.So(auditEvents[0], ShouldResemble, event.Audit{
						CreatedAt:  testTimeMillisInbound,
						StatusCode: int32(http.StatusUnauthorized),
						Path:       "/v1/datasets",
						Method:     http.MethodGet,
						QueryParam: "q1=v1&q2=v2",
					})
				})
			})

			Convey("And a failing audit handler without downstream", func(c C) {
				p, a := createFailingAuditHandler()
				auditHandler := a(nil)

				// execute request and don't expect audit events
				serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

				Convey("Then status Unauthorised and empty body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusUnauthorized)
					b, err := ioutil.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, []byte{})
				})
			})
		})

		Convey("An incoming request with a valid Florence Token", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
			So(err, ShouldBeNil)
			req.Header.Set(dprequest.FlorenceHeaderKey, testFlorenceToken)
			w := httptest.NewRecorder()

			Convey("And a valid audit handler with successful downstream", func(c C) {
				p, a := createValidAuditHandler()
				auditHandler := a(testHandler(http.StatusOK, testBody, c))

				// execute request and wait for 2 audit events
				auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

				Convey("Then status OK and expected body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusOK)
					b, err := ioutil.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, testBody)
				})

				Convey("The expected audit events are sent before and after proxying the call", func() {
					c.So(auditEvents[0], ShouldResemble, event.Audit{
						CreatedAt:  testTimeMillisInbound,
						Path:       "/v1/datasets",
						Method:     http.MethodGet,
						QueryParam: "q1=v1&q2=v2",
						Identity:   testIdentity,
					})
					c.So(auditEvents[1], ShouldResemble, event.Audit{
						CreatedAt:  testTimeMillisOutbound,
						StatusCode: int32(http.StatusOK),
						Path:       "/v1/datasets",
						Method:     http.MethodGet,
						QueryParam: "q1=v1&q2=v2",
						Identity:   testIdentity,
					})
				})
			})

			Convey("And a failing audit handler with successful downstream", func(c C) {
				p, a := createFailingAuditHandler()
				auditHandler := a(testHandler(http.StatusOK, testBody, c))

				// execute request and don't expect audit events
				serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

				Convey("Then status 500 and empty body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusInternalServerError)
					b, err := ioutil.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, []byte{})
				})
			})
		})

		Convey("An incoming request with a valid Service Auth Token", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
			So(err, ShouldBeNil)
			req.Header.Set(dprequest.AuthHeaderKey, testServiceAuthToken)
			w := httptest.NewRecorder()

			Convey("And a valid audit handler with successful downstream", func(c C) {
				p, a := createValidAuditHandler()
				auditHandler := a(testHandler(http.StatusOK, testBody, c))

				// execute request and wait for 2 audit events
				auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

				Convey("Then status OK and expected body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusOK)
					b, err := ioutil.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, testBody)
				})

				Convey("The expected audit events are sent before and after proxying the call", func() {
					c.So(auditEvents[0], ShouldResemble, event.Audit{
						CreatedAt:  testTimeMillisInbound,
						Path:       "/v1/datasets",
						Method:     http.MethodGet,
						QueryParam: "q1=v1&q2=v2",
						Identity:   testIdentity,
					})
					c.So(auditEvents[1], ShouldResemble, event.Audit{
						CreatedAt:  testTimeMillisOutbound,
						StatusCode: int32(http.StatusOK),
						Path:       "/v1/datasets",
						Method:     http.MethodGet,
						QueryParam: "q1=v1&q2=v2",
						Identity:   testIdentity,
					})
				})
			})

			Convey("And a failing audit handler with successful downstream", func(c C) {
				p, a := createFailingAuditHandler()
				auditHandler := a(testHandler(http.StatusOK, testBody, c))

				// execute request and don't expect audit events
				serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

				Convey("Then status 500 and empty body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusInternalServerError)
					b, err := ioutil.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, []byte{})
				})
			})
		})

		Convey("An incoming request with all headers", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
			So(err, ShouldBeNil)
			req.Header.Set(dprequest.AuthHeaderKey, testServiceAuthToken)
			req.Header.Set(dprequest.FlorenceHeaderKey, testFlorenceToken)
			req.Header.Set(dprequest.CollectionIDHeaderKey, testCollectionID)
			req = req.WithContext(context.WithValue(req.Context(), dprequest.RequestIdKey, testRequestID))
			w := httptest.NewRecorder()

			Convey("And a valid audit handler with successful downstream", func(c C) {
				p, a := createValidAuditHandler()
				auditHandler := a(testHandler(http.StatusOK, testBody, c))

				// execute request and wait for 2 audit events
				auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

				Convey("Then status OK and expected body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusOK)
					b, err := ioutil.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, testBody)
				})

				Convey("The expected audit events are sent before and after proxying the call with collectionID and requestID values", func() {
					c.So(auditEvents[0], ShouldResemble, event.Audit{
						CreatedAt:    testTimeMillisInbound,
						Path:         "/v1/datasets",
						Method:       http.MethodGet,
						QueryParam:   "q1=v1&q2=v2",
						Identity:     testIdentity,
						CollectionID: testCollectionID,
						RequestID:    testRequestID,
					})
					c.So(auditEvents[1], ShouldResemble, event.Audit{
						CreatedAt:    testTimeMillisOutbound,
						StatusCode:   int32(http.StatusOK),
						Path:         "/v1/datasets",
						Method:       http.MethodGet,
						QueryParam:   "q1=v1&q2=v2",
						Identity:     testIdentity,
						CollectionID: testCollectionID,
						RequestID:    testRequestID,
					})
				})
			})
		})
	})
}

func TestAuditHandler(t *testing.T) {

	Convey("Given deterministic inbound and outbound timestamps, and an incoming request with valid Florence and Service tokens", t, func(c C) {

		isInbound := true
		middleware.Now = func() time.Time {
			if isInbound {
				isInbound = false
				return testTimeInbound
			}
			return testTimeOutbound
		}

		req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
		So(err, ShouldBeNil)
		req.Header.Set(dprequest.FlorenceHeaderKey, testFlorenceToken)
		req.Header.Set(dprequest.AuthHeaderKey, testServiceAuthToken)
		w := httptest.NewRecorder()

		Convey("And a valid audit handler with unsuccessful (Forbidden) downstream", func(c C) {
			p, a := createValidAuditHandler()
			auditHandler := a(testHandler(http.StatusForbidden, testBody, c))

			// execute request and wait for 2 audit events
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

			Convey("Then status Forbidden and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusForbidden)
				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, testBody)
			})

			Convey("The expected audit events are sent before and after proxying the call", func(c C) {
				c.So(auditEvents[0], ShouldResemble, event.Audit{
					CreatedAt:  testTimeMillisInbound,
					Path:       "/v1/datasets",
					Method:     http.MethodGet,
					QueryParam: "q1=v1&q2=v2",
					Identity:   testIdentity,
				})
				c.So(auditEvents[1], ShouldResemble, event.Audit{
					CreatedAt:  testTimeMillisOutbound,
					StatusCode: int32(http.StatusForbidden),
					Path:       "/v1/datasets",
					Method:     http.MethodGet,
					QueryParam: "q1=v1&q2=v2",
					Identity:   testIdentity,
				})
			})
		})

		Convey("And a failing audit handler with unsuccessful (Forbidden) downstream", func(c C) {
			p, a := createFailingAuditHandler()
			auditHandler := a(testHandler(http.StatusForbidden, testBody, c))

			// execute request and don't expect audit events
			serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

			Convey("Then status 500 and empty body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusInternalServerError)
				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, []byte{})
			})
		})

		Convey("And an audit handler that fails only on the outbound auditing with unsuccessful (Forbidden) downstream", func(c C) {
			cliMock := createHTTPClientMock(http.StatusOK, testIdentityResponse)
			inboundAuditEvent := event.Audit{
				CreatedAt: testTimeMillisInbound,
				Path:      "/v1/datasets",
				Method:    http.MethodGet,
			}
			nMarshalCall := 0
			failingMarshaller := &eventmock.MarshallerMock{
				MarshalFunc: func(s interface{}) ([]byte, error) {
					if nMarshalCall == 0 {
						nMarshalCall++
						b, err := schema.AuditEvent.Marshal(inboundAuditEvent)
						c.So(err, ShouldBeNil)
						return b, nil
					}
					return []byte{}, errMarshal
				},
			}
			p := kafkatest.NewMessageProducer(true)
			a := event.NewAvroProducer(p.Channels().Output, failingMarshaller)
			enableZebedeeAudit := true
			auditHandler := middleware.AuditHandler(a, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, nil)(testHandler(http.StatusForbidden, testBody, c))

			// execute request and expect only 1 audit event
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 1)

			Convey("Then status 500 and empty body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusInternalServerError)
				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, []byte{})
			})

			Convey("The expected audit event is sent before proxying the call", func(c C) {
				c.So(auditEvents[0], ShouldResemble, inboundAuditEvent)
			})
		})
	})
}

func TestAuditIgnoreSkip(t *testing.T) {

	Convey("Given an incoming request to an ignored path", t, func(c C) {
		req, err := http.NewRequest(http.MethodGet, "/ping", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		Convey("And a valid audit handler without downstream", func(c C) {

			p, a := createValidAuditHandler()
			auditHandler := a(testHandler(http.StatusForbidden, testBody, c))

			// execute request and don't wait for audit events
			serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

			Convey("Then status Forbidden and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusForbidden)
				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, testBody)
			})
		})
	})

	Convey("Given an incoming request to a path for which identity check needs to be skipped", t, func(c C) {
		req, err := http.NewRequest(http.MethodGet, "/login", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		Convey("And a valid audit handler without downstream", func(c C) {

			p, a := createValidAuditHandler()
			auditHandler := a(testHandler(http.StatusForbidden, testBody, c))

			// execute request and wait for 2 audit events
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

			Convey("Then status Forbidden and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusForbidden)
				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, testBody)
			})

			Convey("The expected audit events are sent before and after, without identity", func() {
				c.So(auditEvents[0].Identity, ShouldResemble, "")
				c.So(auditEvents[1].Identity, ShouldResemble, "")
			})

		})
	})
}

func TestSkipZebedeeAudit(t *testing.T) {

	Convey("Given an audit handler configured to not audit zebedee requests", t, func(c C) {

		cliMock := createHTTPClientMock(http.StatusOK, testIdentityResponse)
		p := kafkatest.NewMessageProducer(true)
		auditProducer := event.NewAvroProducer(p.Channels().Output, schema.AuditEvent)
		enableZebedeeAudit := false
		routerMock := &mock.RouterMock{
			MatchFunc: func(req *http.Request, match *mux.RouteMatch) bool {
				match.MatchErr = mux.ErrNotFound
				return true
			},
		}
		auditMiddleware := middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, routerMock)
		auditHandler := auditMiddleware(testHandler(http.StatusOK, testBody, c))

		Convey("When the handler receives a Zebedee request", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/data", nil)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			// execute request and don't wait for audit events
			serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

			Convey("Then status OK and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusOK)
				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, testBody)
			})
		})
	})

	Convey("Given an audit handler configured to not audit zebedee requests", t, func(c C) {
		isInbound := true
		middleware.Now = func() time.Time {
			if isInbound {
				isInbound = false
				return testTimeInbound
			}
			return testTimeOutbound
		}

		cliMock := createHTTPClientMock(http.StatusOK, testIdentityResponse)
		p := kafkatest.NewMessageProducer(true)
		auditProducer := event.NewAvroProducer(p.Channels().Output, schema.AuditEvent)
		enableZebedeeAudit := false
		routerMock := &mock.RouterMock{
			MatchFunc: func(req *http.Request, match *mux.RouteMatch) bool {
				return true
			},
		}
		auditMiddleware := middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, routerMock)
		auditHandler := auditMiddleware(testHandler(http.StatusOK, testBody, c))

		Convey("When the handler receives a request for a known route (not zebedee)", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", nil)
			req.Header.Set(dprequest.FlorenceHeaderKey, testFlorenceToken)
			req.Header.Set(dprequest.AuthHeaderKey, testServiceAuthToken)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			// execute request and wait for audit event
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

			Convey("Then status OK and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusOK)
				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, testBody)
			})

			Convey("The expected audit events are sent before and after proxying the call", func() {
				c.So(auditEvents[0], ShouldResemble, event.Audit{
					CreatedAt:  testTimeMillisInbound,
					Path:       "/v1/datasets",
					Method:     http.MethodGet,
					QueryParam: "q1=v1&q2=v2",
					Identity:   testIdentity,
				})
				c.So(auditEvents[1], ShouldResemble, event.Audit{
					CreatedAt:  testTimeMillisOutbound,
					StatusCode: int32(http.StatusOK),
					Path:       "/v1/datasets",
					Method:     http.MethodGet,
					QueryParam: "q1=v1&q2=v2",
					Identity:   testIdentity,
				})
			})
		})
	})
}

func TestShallSkipIdentity(t *testing.T) {
	skipIdentityPaths := []string{"/password", "/login", "/hierarchies"}

	for _, p := range skipIdentityPaths {
		versionedPath := "/v1" + p
		Convey(fmt.Sprintf("Should skip identity check for versioned %s requests", versionedPath), t, func() {
			So(middleware.ShallSkipIdentity("v1", versionedPath), ShouldBeTrue)
		})

		Convey(fmt.Sprintf("Should skip identity check for non versioned %s requests", p), t, func() {
			So(middleware.ShallSkipIdentity("v1", p), ShouldBeTrue)
		})
	}
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
