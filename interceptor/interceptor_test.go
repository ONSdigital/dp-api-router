package interceptor

import (
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testDomain = "https://beta.ons.gov.uk"

func TestUnitInterceptor(t *testing.T) {
	Convey("test interceptor correctly updates a href in links subdoc", t, func() {
		testJSON := `{"links":{"self":{"href":"/datasets/12345"}}}`
		respW := httptest.NewRecorder()

		w := writer{respW, testDomain}

		n, err := w.Write([]byte(testJSON))
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 72)

		So(respW.Body.String(), ShouldEqual, `{"links":{"self":{"href":"https://api.beta.ons.gov.uk/datasets/12345"}}}`)
	})

	Convey("test interceptor correctly updates a href in downloads subdoc", t, func() {
		testJSON := `{"downloads":{"csv":{"href":"http://localhost:22000/myfile.csv"}}}`
		respW := httptest.NewRecorder()

		w := writer{respW, testDomain}

		n, err := w.Write([]byte(testJSON))
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 76)

		So(respW.Body.String(), ShouldEqual, `{"downloads":{"csv":{"href":"https://download.beta.ons.gov.uk/myfile.csv"}}}`)
	})

	Convey("test interceptor correctly updates a href in dimensions subdoc", t, func() {
		testJSON := `{"dimensions":[{"href":"http://localhost:23000/codelists/1234567"}]}`
		respW := httptest.NewRecorder()

		w := writer{respW, testDomain}

		n, err := w.Write([]byte(testJSON))
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 73)

		So(respW.Body.String(), ShouldEqual, `{"dimensions":[{"href":"https://api.beta.ons.gov.uk/codelists/1234567"}]}`)
	})

	Convey("test interceptor correctly updates a nested links document", t, func() {
		testJSON := `{"items":[{"links":{"self":{"href":"/datasets/12345"}}}]}`
		respW := httptest.NewRecorder()

		w := writer{respW, testDomain}

		n, err := w.Write([]byte(testJSON))
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 84)

		So(respW.Body.String(), ShouldEqual, `{"items":[{"links":{"self":{"href":"https://api.beta.ons.gov.uk/datasets/12345"}}}]}`)
	})

	Convey("test interceptor correctly updates a nested array of links", t, func() {
		testJSON := `{"links":{"instances":[{"href":"/datasets/12345"}]}}`
		respW := httptest.NewRecorder()

		w := writer{respW, testDomain}

		n, err := w.Write([]byte(testJSON))
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 79)

		So(respW.Body.String(), ShouldEqual, `{"links":{"instances":[{"href":"https://api.beta.ons.gov.uk/datasets/12345"}]}}`)
	})

	Convey("test interceptor returns an error on write if bytes are not valid json", t, func() {
		b := `not valid json sorry ¯\_(ツ)_/¯`
		respW := httptest.NewRecorder()

		w := writer{respW, testDomain}

		n, err := w.Write([]byte(b))
		So(err, ShouldNotBeNil)
		So(n, ShouldEqual, 0)
	})
}
