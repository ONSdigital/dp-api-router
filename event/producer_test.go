package event_test

import (
	"testing"

	"github.com/ONSdigital/dp-api-router/event"
	"github.com/ONSdigital/dp-api-router/event/mock"
	"github.com/ONSdigital/dp-api-router/schema"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var errMarshal = errors.New("Marshal error")

var testAuditEvent = &event.Audit{
	CreatedAt:    testTimeMillis,
	RequestID:    "myRequest",
	Identity:     "myIdentity",
	CollectionID: "myCollection",
	Path:         "myPath",
	Method:       "myMethod",
	StatusCode:   200,
	QueryParam:   "myQueryParam",
}

func TestAvroProducer(t *testing.T) {
	Convey("Given a successful message producer mock", t, func() {
		// channel to capture messages sent.
		outputChannel := make(chan []byte, 1)

		// bytes to send
		avroBytes := []byte("hello world")

		// mock that represents a marshaller
		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return avroBytes, nil
			},
		}

		Convey("when Audit is called with a nil event", func() {
			// eventProducer under test
			eventProducer := event.NewAvroProducer(outputChannel, marshallerMock)
			err := eventProducer.Audit(nil)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, "event required but was nil")
			})

			Convey("and marshaller is never called", func() {
				So(len(marshallerMock.MarshalCalls()), ShouldEqual, 0)
			})
		})

		Convey("When Audit is called on the event producer", func() {
			// eventProducer under test
			eventProducer := event.NewAvroProducer(outputChannel, schema.AuditEvent)
			err := eventProducer.Audit(testAuditEvent)

			Convey("The expected event is available on the output channel", func() {
				So(err, ShouldBeNil)

				messageBytes := <-outputChannel
				close(outputChannel)
				sentEvent := unmarshal(messageBytes)
				So(sentEvent, ShouldResemble, testAuditEvent)
			})
		})
	})

	Convey("Given a message producer mock that fails to marshall", t, func() {
		// mock that represents a marshaller
		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return nil, errMarshal
			},
		}

		// eventProducer under test, without out channel because nothing is expected to be sent
		eventProducer := event.NewAvroProducer(nil, marshallerMock)

		Convey("When Audit is called on the event producer", func() {
			err := eventProducer.Audit(testAuditEvent)

			Convey("The expected error is returned", func() {
				So(err, ShouldResemble, errMarshal)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func unmarshal(bytes []byte) *event.Audit {
	observationEvent := &event.Audit{}
	err := schema.AuditEvent.Unmarshal(bytes, observationEvent)
	So(err, ShouldBeNil)
	return observationEvent
}
