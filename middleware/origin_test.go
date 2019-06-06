package middleware

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
)

const (
	whiteListedOrigin1 = "i-r-a-good-place.com"
	whiteListedOrigin2 = "is-a-better-place.fam"
)

var (
	dummyHandler       = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
	whiteListedOrigins = []string{whiteListedOrigin1, whiteListedOrigin2}
)

func TestOriginHandler(t *testing.T) {

	Convey("origin handler should wrap another handler", t, func() {
		handler := SetAllowOriginHeader([]string{"origin-is-allowed.com"})
		wrapped := handler(dummyHandler)
		So(wrapped, ShouldHaveSameTypeAs, dummyHandler)
	})

	Convey("origin handler should serve the request where the origin is allowed", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("Origin", whiteListedOrigin1)
		w := httptest.NewRecorder()

		handler := SetAllowOriginHeader(whiteListedOrigins)
		wrapped := handler(dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)
		So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, whiteListedOrigin1)

	})

	Convey("origin handler should return 401 unauthorised where origin is not allowed", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("Origin", whiteListedOrigin1)
		w := httptest.NewRecorder()

		handler := SetAllowOriginHeader([]string{})
		wrapped := handler(dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 401)
	})

}
