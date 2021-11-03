package interceptor

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testDomain = "https://beta.ons.gov.uk/v1"
const testContext = "context.json"

type dummyRT struct {
	testJSON string
}

func (t dummyRT) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	_ = req // shut some linters up
	resp = httptest.NewRecorder().Result()
	resp.Body = ioutil.NopCloser(strings.NewReader(t.testJSON))
	return
}

var _ http.RoundTripper = dummyRT{}

func TestUnitInterceptor(t *testing.T) {

	Convey("test interceptor doesn't throw an error for an empty response", t, func() {
		testJSON := ``
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 0)
	})

	Convey("test interceptor doesn't change an already correct link", t, func() {
		testJSON := `{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, `{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}`+"\n")
	})

	Convey("test interceptor correctly updates a href in links subdoc", t, func() {
		testJSON := `{"links":{"self":{"href":"/datasets/12345"}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 76)
		So(string(b), ShouldEqual, `{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}`+"\n")
	})

	Convey("test interceptor correctly inserts context", t, func() {
		testJSON := `{"links":{"self":{"href":"/datasets/12345"}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, testContext, transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 102)
		So(string(b), ShouldEqual, `{"@context":"context.json","links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}`+"\n")
	})

	Convey("test interceptor correctly updates a href in downloads subdoc on a nested path", t, func() {
		testJSON := `{"downloads":{"csv":{"href":"http://localhost:22000/myfile.csv"}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets/1234"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 77)
		So(string(b), ShouldEqual, `{"downloads":{"csv":{"href":"https://download.beta.ons.gov.uk/myfile.csv"}}}`+"\n")
	})

	Convey("test interceptor correctly updates a href in dimensions subdoc", t, func() {
		testJSON := `{"dimensions":[{"href":"http://localhost:23000/code-lists/1234567"}]}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/code-lists"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 78)
		So(string(b), ShouldEqual, `{"dimensions":[{"href":"https://api.beta.ons.gov.uk/v1/code-lists/1234567"}]}`+"\n")
	})

	Convey("test interceptor correctly updates a nested links document", t, func() {
		testJSON := `{"items":[{"links":{"self":{"href":"/datasets/12345"}}}]}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 88)
		So(string(b), ShouldEqual, `{"items":[{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}]}`+"\n")
	})

	Convey("test interceptor correctly updates a nested array of links", t, func() {
		testJSON := `{"links":{"instances":[{"href":"/datasets/12345"}]}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 83)
		So(string(b), ShouldEqual, `{"links":{"instances":[{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}]}}`+"\n")
	})

	Convey("test interceptor correctly updates a nested dimension href", t, func() {
		testJSON := `{"dimensions":{"time":{"option":{"href":"/datasets/time"}}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 91)
		So(string(b), ShouldEqual, `{"dimensions":{"time":{"option":{"href":"https://api.beta.ons.gov.uk/v1/datasets/time"}}}}`+"\n")
	})

	Convey("test query parameters are parsed correctly on response body rewrite", t, func() {
		testJSON := `{"dimensions":{"time":{"option":{"href":"/datasets/time?hello=world&mobile=phone"}}}}`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 116)
		So(string(b), ShouldEqual, `{"dimensions":{"time":{"option":{"href":"https://api.beta.ons.gov.uk/v1/datasets/time?hello=world&mobile=phone"}}}}`+"\n")
	})

	Convey("test interceptor correctly updates a href in links subdocs within an array", t, func() {
		testJSON := `[{"links":{"self":{"href":"/datasets/12345"}}}, {"links":{"self":{"href":"/datasets/12345"}}}]`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, 154)
		So(string(b), ShouldEqual, `[{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}},{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}]`+"\n")
	})

	Convey("test interceptor correctly ignores non json and non map object that is 'maxBodyLengthToLog - 1'", t, func() {
		testJSON := "A"
		for i := 0; i < (maxBodyLengthToLog - 2); i++ {
			testJSON += "n"
		}
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, maxBodyLengthToLog-1)
		So(string(b), ShouldEqual, testJSON)
	})

	Convey("test interceptor correctly ignores non json and non map object that is 'maxBodyLengthToLog'", t, func() {
		testJSON := "B"
		for i := 0; i < (maxBodyLengthToLog - 1); i++ {
			testJSON += "n"
		}
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, maxBodyLengthToLog)
		So(string(b), ShouldEqual, testJSON)
	})

	Convey("test interceptor correctly ignores non json and non map object that is 'maxBodyLengthToLog+1'", t, func() {
		testJSON := "C"
		for i := 0; i < (maxBodyLengthToLog - 1); i++ {
			testJSON += "n"
		}
		testJSON += "1"
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, maxBodyLengthToLog+1)
		So(string(b), ShouldEqual, testJSON)
	})

	Convey("test interceptor correctly handles broken json object, that is not split", t, func() {
		testJSON := `{bla`
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, len(testJSON))
		So(string(b), ShouldEqual, testJSON)
	})

	Convey("test interceptor correctly handles broken json object that will be split", t, func() {
		testJSON := `{bla`
		for i := 0; i < maxBodyLengthToLog; i++ {
			testJSON += "n"
		}
		transp := dummyRT{testJSON}

		t := NewRoundTripper(testDomain, "", transp)

		resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
		So(err, ShouldBeNil)

		b, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		err = resp.Body.Close()
		So(err, ShouldBeNil)
		So(len(b), ShouldEqual, len(testJSON))
		So(string(b), ShouldEqual, testJSON)
	})
}
