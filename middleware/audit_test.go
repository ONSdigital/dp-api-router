package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"

	"github.com/ONSdigital/dp-api-router/middleware/mock"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-api-router/event"
	eventmock "github.com/ONSdigital/dp-api-router/event/mock"
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/ONSdigital/dp-api-router/schema"

	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/dp-kafka/v4/kafkatest"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testVersionPrefix            = "v1"
	testZebedeeURL               = "myZebedeeURL"
	testRequestID                = "myRequest"
	testIdentity                 = "myIdentity"
	testCollectionID             = "myCollection"
	testFlorenceToken            = "myFlorenceToken"
	testJWTFlorenceToken         = "Bearer eyJraWQiOiJBQzBnOXBzZzBwTEJ1Q2Nqa00yVkZEbXlzUlNxNm5KWlNxbkNXd1wvMFk1RT0iLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiI0ZWNjY2NiOS01MDJkLTQxZDEtYWZlMC0xZWRhNWJhNzY2NzAiLCJjb2duaXRvOmdyb3VwcyI6WyJyb2xlLWFkbWluIl0sImlzcyI6Imh0dHBzOlwvXC9jb2duaXRvLWlkcC5ldS13ZXN0LTEuYW1hem9uYXdzLmNvbVwvZXUtd2VzdC0xX0JuN0RhSXU3SCIsImNsaWVudF9pZCI6ImdoMGg4aTdja2N1OGJmMjFwOHIwb2pta2QiLCJvcmlnaW5fanRpIjoiNTViMGY1MGQtOGQyYS00ZThjLWJjN2EtYmMwZjU1OTcxY2ZhIiwiZXZlbnRfaWQiOiI1NWM0YmI3OS1kNjNiLTQ5NzEtYjM4OS04OTlkNmU1MjVkNTUiLCJ0b2tlbl91c2UiOiJhY2Nlc3MiLCJzY29wZSI6ImF3cy5jb2duaXRvLnNpZ25pbi51c2VyLmFkbWluIiwiYXV0aF90aW1lIjoxNjU5NDQ0MDQ2LCJleHAiOjE2NTk0NDQ0MDYsImlhdCI6MTY1OTQ0NDA0NiwianRpIjoiNGE5NWUzYmQtNDk2Yi00YmM0LTk4MTAtOTZhODU4MjgxMTJhIiwidXNlcm5hbWUiOiJhYjMzNDI2My0yYzBiLTRlZDYtODQzNC04Yzg4NDdmZGRhMjgifQ.J5xMdW_ovcOBQWSy9vbxGiF8YeK6p7M_K-miumvPLWFIWv5EwcrjCfCna8Pp3kOk03DOJVRyUWk3hgsrC_sIIJOLKixSTMqM95xUQsP4jd0WSGJF_7rUwBwbfpvj_HLB9hGkzx7LsGAUh2eInjP5oudHHLOPdy9bPwVevg9IBPEfZZ8I9UZ8qnkFCbi2Y29ETOjXUogZL_SCpSs2QQdKG-CuqnWeWVHBvdFfZ-KOPao7ObJsfs_mKFWcro6YKa1J4jxfQhKJjC9qeMz8l7SfcqpeatfmoLogx_wyxyL36319WBXgUthMBb4rpNdFjMerSAv1eq92JkT7dhgtKW3Gpw\n"
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
	config := &kafka.ProducerConfig{
		BrokerAddrs:     []string{"localhost:9092", "localhost:9093"},
        Topic:           "test-topic",
	}
	p, err := kafkatest.NewProducer(context.Background(), config, kafkatest.DefaultProducerConfig)
	if err != nil {
		fmt.Printf("HELLO!!! ")
		fmt.Printf("%+v\n", err)
	}
	auditProducer := event.NewAvroProducer(p.Mock.Channels().Output, schema.AuditEvent)
	enableZebedeeAudit := true
	auth := authorisation.Config{}
	return p.Mock, middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, nil, auth)
}

// utility function to create a producer and an audit handler that fails to marshal and send events
func createFailingAuditHandler() (kafka.IProducer, func(h http.Handler) http.Handler) {
	cliMock := createHTTPClientMock(http.StatusOK, testIdentityResponse)
	failingMarshaller := &eventmock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return []byte{}, errMarshal
		},
	}
	config := &kafka.ProducerConfig{
		BrokerAddrs:     []string{"localhost:9092", "localhost:9093"},
        Topic:           "test-topic",
	}
	p, _ := kafkatest.NewProducer(context.Background(), config, kafkatest.DefaultProducerConfig)
	auditProducer := event.NewAvroProducer(p.Mock.Channels().Output, failingMarshaller)
	enableZebedeeAudit := true
	auth := authorisation.Config{}
	return p.Mock, middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, nil, auth)
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
				Body:       io.NopCloser(bytes.NewReader(body)),
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
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
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
			req, err := http.NewRequest(http.MethodGet, "/v1/data?lang=en\u0026uri=%2Fhealth", http.NoBody)
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
			req, err := http.NewRequest(http.MethodGet, "/v1/data?uri=%wxhealth", http.NoBody)
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
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			Convey("And a valid audit handler without downstream", func(c C) {
				p, a := createValidAuditHandler()
				auditHandler := a(nil)

				// execute request and wait for 1 audit event
				auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 1)

				Convey("Then status Unauthorised and empty body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusUnauthorized)
					b, err := io.ReadAll(w.Body)
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
					b, err := io.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, []byte{})
				})
			})
		})

		Convey("An incoming request with no auth headers but no identy needed", func(c C) {
			req, err := http.NewRequest(http.MethodPut, "/v1/users/self/password", http.NoBody)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			Convey("And a valid audit handler with successful downstream", func(c C) {
				p, a := createValidAuditHandler()
				auditHandler := a(testHandler(http.StatusOK, testBody, c))

				// execute request and wait for 2 audit events
				auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

				Convey("Then status OK and expected body is returned", func(c C) {
					c.So(w.Code, ShouldEqual, http.StatusOK)
					b, err := io.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, testBody)
				})

				Convey("The expected audit events are sent before and after proxying the call", func() {
					c.So(auditEvents[0], ShouldResemble, event.Audit{
						CreatedAt:  testTimeMillisInbound,
						Path:       "/v1/users/self/password",
						Method:     http.MethodPut,
						QueryParam: "",
					})
				})
			})
		})

		Convey("An incoming request with a valid Florence Token", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
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
					b, err := io.ReadAll(w.Body)
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
					b, err := io.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, []byte{})
				})
			})
		})

		Convey("An incoming request with a valid Service Auth Token", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
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
					b, err := io.ReadAll(w.Body)
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
					b, err := io.ReadAll(w.Body)
					So(err, ShouldBeNil)
					c.So(b, ShouldResemble, []byte{})
				})
			})
		})

		Convey("An incoming request with all headers", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
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
					b, err := io.ReadAll(w.Body)
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

		req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
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
				b, err := io.ReadAll(w.Body)
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
				b, err := io.ReadAll(w.Body)
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
			config := &kafka.ProducerConfig{
				BrokerAddrs:     []string{"localhost:9092"},
				Topic:           "test-topic",
			}
			p, _ := kafkatest.NewProducer(context.Background(), config, kafkatest.DefaultProducerConfig)
			a := event.NewAvroProducer(p.Mock.Channels().Output, failingMarshaller)
			enableZebedeeAudit := true
			auth := authorisation.Config{}
			auditHandler := middleware.AuditHandler(a, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, nil, auth)(testHandler(http.StatusForbidden, testBody, c))

			// execute request and expect only 1 audit event
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Mock.Channels().Output, 1)

			Convey("Then status 500 and empty body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusInternalServerError)
				b, err := io.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, []byte{})
			})

			Convey("The expected audit event is sent before proxying the call", func(c C) {
				c.So(auditEvents[0], ShouldResemble, inboundAuditEvent)
			})
		})
	})
}
func TestAuditHandlerJWTFlorenceToken(t *testing.T) {
	Convey("Given deterministic inbound and outbound timestamps, and an incoming request with invalid JWT_Florence and Service tokens", t, func(c C) {
		isInbound := true
		middleware.Now = func() time.Time {
			if isInbound {
				isInbound = false
				return testTimeInbound
			}
			return testTimeOutbound
		}

		req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
		So(err, ShouldBeNil)
		req.Header.Set(dprequest.FlorenceHeaderKey, testJWTFlorenceToken)
		req.Header.Set(dprequest.AuthHeaderKey, testServiceAuthToken)
		w := httptest.NewRecorder()

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
			config := &kafka.ProducerConfig{
				BrokerAddrs:     []string{"localhost:9092"},
				Topic:           "test-topic",
			}
			p, _ := kafkatest.NewProducer(context.Background(), config, kafkatest.DefaultProducerConfig)
			a := event.NewAvroProducer(p.Mock.Channels().Output, failingMarshaller)
			enableZebedeeAudit := true
			auth := authorisation.Config{}
			auditHandler := middleware.AuditHandler(a, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, nil, auth)(testHandler(http.StatusForbidden, testBody, c))

			// execute request and expect only 1 audit event
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Mock.Channels().Output, 1)
			Convey("Then status 500 and empty body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusInternalServerError)
				b, err := io.ReadAll(w.Body)
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
		req, err := http.NewRequest(http.MethodGet, "/ping", http.NoBody)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		Convey("And a valid audit handler without downstream", func(c C) {
			p, a := createValidAuditHandler()
			auditHandler := a(testHandler(http.StatusForbidden, testBody, c))

			// execute request and don't wait for audit events
			serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 0)

			Convey("Then status Forbidden and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusForbidden)
				b, err := io.ReadAll(w.Body)
				So(err, ShouldBeNil)
				c.So(b, ShouldResemble, testBody)
			})
		})
	})

	Convey("Given an incoming request to a path for which identity check needs to be skipped", t, func(c C) {
		req, err := http.NewRequest(http.MethodGet, "/login", http.NoBody)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		Convey("And a valid audit handler without downstream", func(c C) {
			p, a := createValidAuditHandler()
			auditHandler := a(testHandler(http.StatusForbidden, testBody, c))

			// execute request and wait for 2 audit events
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Channels().Output, 2)

			Convey("Then status Forbidden and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusForbidden)
				b, err := io.ReadAll(w.Body)
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
		config := &kafka.ProducerConfig{
			BrokerAddrs:     []string{"localhost:9092"},
			Topic:           "test-topic",
		}
		p, _ := kafkatest.NewProducer(context.Background(), config, kafkatest.DefaultProducerConfig)
		auditProducer := event.NewAvroProducer(p.Mock.Channels().Output, schema.AuditEvent)
		enableZebedeeAudit := false
		auth := authorisation.Config{}
		routerMock := &mock.RouterMock{
			MatchFunc: func(req *http.Request, match *mux.RouteMatch) bool {
				match.MatchErr = mux.ErrNotFound
				return true
			},
		}
		auditMiddleware := middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, routerMock, auth)
		auditHandler := auditMiddleware(testHandler(http.StatusOK, testBody, c))

		Convey("When the handler receives a Zebedee request", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/data", http.NoBody)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			// execute request and don't wait for audit events
			serveAndCaptureAudit(c, w, req, auditHandler, p.Mock.Channels().Output, 0)

			Convey("Then status OK and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusOK)
				b, err := io.ReadAll(w.Body)
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
		config := &kafka.ProducerConfig{
			BrokerAddrs:     []string{"localhost:9092"},
			Topic:           "test-topic",
		}
		p, _ := kafkatest.NewProducer(context.Background(), config, kafkatest.DefaultProducerConfig)
		auditProducer := event.NewAvroProducer(p.Mock.Channels().Output, schema.AuditEvent)
		enableZebedeeAudit := false
		auth := authorisation.Config{}
		routerMock := &mock.RouterMock{
			MatchFunc: func(req *http.Request, match *mux.RouteMatch) bool {
				return true
			},
		}
		auditMiddleware := middleware.AuditHandler(auditProducer, cliMock, testZebedeeURL, testVersionPrefix, enableZebedeeAudit, routerMock, auth)
		auditHandler := auditMiddleware(testHandler(http.StatusOK, testBody, c))

		Convey("When the handler receives a request for a known route (not zebedee)", func(c C) {
			req, err := http.NewRequest(http.MethodGet, "/v1/datasets?q1=v1&q2=v2", http.NoBody)
			req.Header.Set(dprequest.FlorenceHeaderKey, testFlorenceToken)
			req.Header.Set(dprequest.AuthHeaderKey, testServiceAuthToken)
			So(err, ShouldBeNil)
			w := httptest.NewRecorder()

			// execute request and wait for audit event
			auditEvents := serveAndCaptureAudit(c, w, req, auditHandler, p.Mock.Channels().Output, 2)

			Convey("Then status OK and expected body is returned", func(c C) {
				c.So(w.Code, ShouldEqual, http.StatusOK)
				b, err := io.ReadAll(w.Body)
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
func serveAndCaptureAudit(c C, w http.ResponseWriter, req *http.Request, auditHandler http.Handler, outChan chan kafka.BytesMessage, numExpectedMessages int) (auditEvents []event.Audit) {
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
func captureAuditEvent(c C, outChan chan kafka.BytesMessage) event.Audit {
	message := <-outChan
	auditEvent := event.Audit{}
	err := schema.AuditEvent.Unmarshal(message.Value, &auditEvent)
	c.So(err, ShouldBeNil)
	return auditEvent
}
