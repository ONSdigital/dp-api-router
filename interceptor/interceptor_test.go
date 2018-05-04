package interceptor

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testDomain = "https://beta.ons.gov.uk"

type dummyRT struct {
	testJSON string
}

func (t dummyRT) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp = httptest.NewRecorder().Result()
	resp.Body = ioutil.NopCloser(strings.NewReader(t.testJSON))
	return
}

var _ http.RoundTripper = dummyRT{}

func TestUnitInterceptor(t *testing.T) {
	Convey("test interceptor correctly updates a href in links subdoc", t, func() {
		testJSON := `{"links":{"self":{"href":"/datasets/12345"}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, transp)

		resp, err := t.RoundTrip(nil)
		So(err, ShouldBeNil)

		b, _ := ioutil.ReadAll(resp.Body)

		So(len(b), ShouldEqual, 72)

		So(string(b), ShouldEqual, `{"links":{"self":{"href":"https://api.beta.ons.gov.uk/datasets/12345"}}}`)
	})

	Convey("test interceptor correctly updates a href in downloads subdoc", t, func() {
		testJSON := `{"downloads":{"csv":{"href":"http://localhost:22000/myfile.csv"}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, transp)

		resp, err := t.RoundTrip(nil)
		So(err, ShouldBeNil)

		b, _ := ioutil.ReadAll(resp.Body)

		So(len(b), ShouldEqual, 76)

		So(string(b), ShouldEqual, `{"downloads":{"csv":{"href":"https://download.beta.ons.gov.uk/myfile.csv"}}}`)
	})

	Convey("test interceptor correctly updates a href in dimensions subdoc", t, func() {
		testJSON := `{"dimensions":[{"href":"http://localhost:23000/codelists/1234567"}]}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, transp)

		resp, err := t.RoundTrip(nil)
		So(err, ShouldBeNil)

		b, _ := ioutil.ReadAll(resp.Body)
		So(len(b), ShouldEqual, 73)

		So(string(b), ShouldEqual, `{"dimensions":[{"href":"https://api.beta.ons.gov.uk/codelists/1234567"}]}`)
	})

	Convey("test interceptor correctly updates a nested links document", t, func() {
		testJSON := `{"items":[{"links":{"self":{"href":"/datasets/12345"}}}]}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, transp)

		resp, err := t.RoundTrip(nil)
		So(err, ShouldBeNil)

		b, _ := ioutil.ReadAll(resp.Body)
		So(len(b), ShouldEqual, 84)

		So(string(b), ShouldEqual, `{"items":[{"links":{"self":{"href":"https://api.beta.ons.gov.uk/datasets/12345"}}}]}`)
	})

	Convey("test interceptor correctly updates a nested array of links", t, func() {
		testJSON := `{"links":{"instances":[{"href":"/datasets/12345"}]}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, transp)

		resp, err := t.RoundTrip(nil)
		So(err, ShouldBeNil)

		b, _ := ioutil.ReadAll(resp.Body)
		So(len(b), ShouldEqual, 79)

		So(string(b), ShouldEqual, `{"links":{"instances":[{"href":"https://api.beta.ons.gov.uk/datasets/12345"}]}}`)
	})

	Convey("test interceptor returns an error on write if bytes are not valid json", t, func() {
		b := `not valid json sorry ¯\_(ツ)_/¯`
		transp := dummyRT{b}

		t := NewRoundTripper(testDomain, transp)

		_, err := t.RoundTrip(nil)
		So(err, ShouldNotBeNil)
	})
}
