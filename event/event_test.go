package event_test

import (
	"testing"
	"time"

	event "github.com/ONSdigital/dp-api-router/event"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTime             = time.Date(2020, time.April, 26, 7, 5, 52, 123000000, time.UTC)
	testTimeMillis int64 = 1587884752123
)

func TestAuditEvent(t *testing.T) {
	Convey("Given an Audit struct with a valid CreatedAt time, in millisencods since unix time reference", t, func() {
		auditEvent := event.Audit{
			CreatedAt: testTimeMillis,
		}
		Convey("Then CreatedAtTime returns the expected time struct, in UTC", func() {
			So(auditEvent.CreatedAtTime(), ShouldResemble, testTime)
		})
	})

	Convey("A time struct is correctly translated to the number of milliseconds since unix time reference", t, func() {
		So(event.CreatedAtMillis(testTime), ShouldEqual, testTimeMillis)
	})
}
