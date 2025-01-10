package deprecation

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRouter(t *testing.T) {
	const (
		testResponse          = "hello world"
		deprecatedLink        = "https://deprecated.example.com/some_link"
		deprecatedMessage     = "some message deprecated"
		deprecatedDeprecation = "Mon, 06 Jan 2025 01:11:28 GMT\n"
		deprecatedSunset      = "Tue, 07 Jan 2025 07:41:11 GMT\n"
		outageLink            = "https://outage.example.com/some_link"
		outageMessage         = "some message outage"
		outageDeprecation     = "Thu, 09 Jan 2025 14:23:59 GMT\n"
		outageSunset          = "Fri, 10 Jan 2025 19:00:02 GMT\n"
		inactiveLink          = "https://inactive.example.com/some_link"
		inactiveMessage       = "some message inactive"
		inactiveDeprecation   = "Sat, 11 Jan 2025 04:10:30 GMT\n"
		inactiveSunset        = "Sun, 12 Jan 2025 22:47:12 GMT\n"
	)

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testResponse))
	})

	Convey("Given a deprecation Router", t, func() {
		deprecations := []Deprecation{
			{
				Paths:   []string{"/deprecated"},
				Date:    deprecatedDeprecation,
				Link:    deprecatedLink,
				Message: deprecatedMessage,
				Sunset:  deprecatedSunset,
				Outages: nil,
			},
			{
				Paths:   []string{"/outage"},
				Date:    outageDeprecation,
				Link:    outageLink,
				Message: outageMessage,
				Sunset:  outageSunset,
				Outages: []Outage{
					{Start: time.Now().Add(-3 * time.Hour), End: time.Now().Add(-2 * time.Hour)},
					{Start: time.Now().Add(-time.Hour), End: time.Now().Add(time.Hour)},
					{Start: time.Now().Add(2 * time.Hour), End: time.Now().Add(3 * time.Hour)},
				},
			},
			{
				Paths:   []string{"/inactive"},
				Date:    inactiveDeprecation,
				Link:    inactiveLink,
				Message: inactiveMessage,
				Sunset:  inactiveSunset,
				Outages: []Outage{
					{Start: time.Now().Add(-3 * time.Hour), End: time.Now().Add(-2 * time.Hour)},
					{Start: time.Now().Add(2 * time.Hour), End: time.Now().Add(3 * time.Hour)},
				},
			},
		}
		router := Router(deprecations)(baseHandler)
		So(router, ShouldNotBeNil)
		So(router, ShouldNotEqual, baseHandler)

		Convey("With a request to a non routed path", func() {
			req := httptest.NewRequest("GET", "http://example.com/nonmatched", nil)
			Convey("When the request is made to the router", func() {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				resp := w.Result()
				Convey("Then the response should be handled by the underlying handler", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					body, err := io.ReadAll(resp.Body)
					So(err, ShouldBeNil)
					So(body, ShouldResemble, []byte(testResponse))
				})
				Convey("And the deprecation headers should not be returned", func() {
					headers := resp.Header
					So(headers.Get("Deprecation"), ShouldBeEmpty)
					So(headers.Get("Link"), ShouldBeEmpty)
					So(headers.Get("Sunset"), ShouldBeEmpty)
				})
			})
		})

		Convey("With a request to a deprecated path but no outage", func() {
			req := httptest.NewRequest("GET", "http://example.com/deprecated", nil)
			Convey("When the request is made to the router", func() {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				resp := w.Result()
				Convey("Then the response should be handled by the underlying handler", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					body, err := io.ReadAll(resp.Body)
					So(err, ShouldBeNil)
					So(body, ShouldResemble, []byte(testResponse))
				})
				Convey("And the deprecation headers should be returned", func() {
					headers := resp.Header
					So(headers.Get("Deprecation"), ShouldNotBeEmpty)
					So(headers.Get("Deprecation"), ShouldEqual, deprecatedDeprecation)
					So(headers.Get("Link"), ShouldNotBeEmpty)
					So(headers.Get("Link"), ShouldEqual, "<"+deprecatedLink+">; rel=\"sunset\"")
					So(headers.Get("Sunset"), ShouldNotBeEmpty)
					So(headers.Get("Sunset"), ShouldEqual, deprecatedSunset)
				})
			})
		})

		Convey("With a request to a deprecated path with a current outage", func() {
			req := httptest.NewRequest("GET", "http://example.com/outage", nil)
			Convey("When the request is made to the router", func() {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				resp := w.Result()
				Convey("Then the response should be handled by the underlying handler", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
					body, err := io.ReadAll(resp.Body)
					So(err, ShouldBeNil)
					So(body, ShouldResemble, []byte(outageMessage+"\n"))
				})
				Convey("And the deprecation headers should be returned", func() {
					headers := resp.Header
					So(headers.Get("Deprecation"), ShouldNotBeEmpty)
					So(headers.Get("Deprecation"), ShouldEqual, outageDeprecation)
					So(headers.Get("Link"), ShouldNotBeEmpty)
					So(headers.Get("Link"), ShouldEqual, "<"+outageLink+">; rel=\"sunset\"")
					So(headers.Get("Sunset"), ShouldNotBeEmpty)
					So(headers.Get("Sunset"), ShouldEqual, outageSunset)
				})
			})
		})

		Convey("With a request to a deprecated path with inactive outages", func() {
			req := httptest.NewRequest("GET", "http://example.com/inactive", nil)
			Convey("When the request is made to the router", func() {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				resp := w.Result()
				Convey("Then the response should be handled by the underlying handler", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					body, err := io.ReadAll(resp.Body)
					So(err, ShouldBeNil)
					So(body, ShouldResemble, []byte(testResponse))
				})
				Convey("And the deprecation headers should be returned", func() {
					headers := resp.Header
					So(headers.Get("Deprecation"), ShouldNotBeEmpty)
					So(headers.Get("Deprecation"), ShouldEqual, inactiveDeprecation)
					So(headers.Get("Link"), ShouldNotBeEmpty)
					So(headers.Get("Link"), ShouldEqual, "<"+inactiveLink+">; rel=\"sunset\"")
					So(headers.Get("Sunset"), ShouldNotBeEmpty)
					So(headers.Get("Sunset"), ShouldEqual, inactiveSunset)
				})
			})
		})
	})

	Convey("Given an empty slice of deprecations", t, func() {
		deprecations := []Deprecation{}
		Convey("When Router is called it should return the underlying handler instead", func() {
			router := Router(deprecations)(baseHandler)
			So(router, ShouldNotBeNil)
			So(router, ShouldEqual, baseHandler)
		})
	})

	Convey("Given an nil slice of deprecations", t, func() {
		var deprecations []Deprecation
		Convey("When Router is called it should return the underlying handler instead", func() {
			router := Router(deprecations)(baseHandler)
			So(router, ShouldNotBeNil)
			So(router, ShouldEqual, baseHandler)
		})
	})
}
