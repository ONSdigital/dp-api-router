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
		testResponse               = "hello world"
		deprecatedLink             = "https://deprecated.example.com/some_link"
		deprecatedMessage          = "some message deprecated"
		deprecatedDeprecation      = "@1736125888" // "Mon, 06 Jan 2025 01:11:28 GMT"
		deprecatedSunset           = "Tue, 07 Jan 2025 07:41:11 GMT"
		outageLink                 = "https://outage.example.com/some_link"
		outageMessage              = "some message outage"
		outageDeprecation          = "@1736432639" // "Thu, 09 Jan 2025 14:23:59 GMT"
		outageSunset               = "Fri, 10 Jan 2025 19:00:02 GMT"
		inactiveLink               = "https://inactive.example.com/some_link"
		inactiveMessage            = "some message inactive"
		inactiveDeprecation        = "@1736568630" // "Sat, 11 Jan 2025 04:10:30 GMT"
		inactiveSunset             = "Sun, 12 Jan 2025 22:47:12 GMT"
		unboundedLink              = "https://unbounded.example.com/some_link"
		unboundedMessage           = "some message unbounded"
		unboundedDeprecation       = "@1737519030" // "Sat, 22 Jan 2025 04:10:30 GMT"
		unboundedSunset            = "Sun, 22 Jan 2025 22:47:12 GMT"
		futureUnboundedLink        = "https://unbounded.example.com/some_link"
		futureUnboundedMessage     = "some message unbounded"
		futureUnboundedDeprecation = "@1737951030" // "Sat, 27 Jan 2025 04:10:30 GMT"
		futureUnboundedSunset      = "Sun, 28 Jan 2025 22:47:12 GMT"
	)

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testResponse))
	})

	Convey("Given a deprecation Router", t, func() {
		deprecations := []Deprecation{
			{
				Paths:    []string{"/deprecated"},
				DateUnix: deprecatedDeprecation,
				Link:     deprecatedLink,
				Message:  deprecatedMessage,
				Sunset:   deprecatedSunset,
				Outages:  nil,
			},
			{
				Paths:    []string{"/outage"},
				DateUnix: outageDeprecation,
				Link:     outageLink,
				Message:  outageMessage,
				Sunset:   outageSunset,
				Outages: []Outage{
					{Start: time.Now().Add(-3 * time.Hour), End: anyToPointer(time.Now().Add(-2 * time.Hour))},
					{Start: time.Now().Add(-time.Hour), End: anyToPointer(time.Now().Add(time.Hour))},
					{Start: time.Now().Add(2 * time.Hour), End: anyToPointer(time.Now().Add(3 * time.Hour))},
				},
			},
			{
				Paths:    []string{"/inactive"},
				DateUnix: inactiveDeprecation,
				Link:     inactiveLink,
				Message:  inactiveMessage,
				Sunset:   inactiveSunset,
				Outages: []Outage{
					{Start: time.Now().Add(-3 * time.Hour), End: anyToPointer(time.Now().Add(-2 * time.Hour))},
					{Start: time.Now().Add(2 * time.Hour), End: anyToPointer(time.Now().Add(3 * time.Hour))},
				},
			},
			{
				Paths:    []string{"/unbounded"},
				DateUnix: unboundedDeprecation,
				Link:     unboundedLink,
				Message:  unboundedMessage,
				Sunset:   unboundedSunset,
				Outages: []Outage{
					{Start: time.Now().Add(-3 * time.Hour), End: anyToPointer(time.Now().Add(-2 * time.Hour))},
					{Start: time.Now().Add(-1 * time.Hour), End: nil},
				},
			},
			{
				Paths:    []string{"/futureunbounded"},
				DateUnix: futureUnboundedDeprecation,
				Link:     futureUnboundedLink,
				Message:  futureUnboundedMessage,
				Sunset:   futureUnboundedSunset,
				Outages: []Outage{
					{Start: time.Now().Add(-3 * time.Hour), End: anyToPointer(time.Now().Add(-2 * time.Hour))},
					{Start: time.Now().Add(1 * time.Hour), End: nil},
					{Start: time.Now().Add(2 * time.Hour), End: anyToPointer(time.Now().Add(3 * time.Hour))},
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

		Convey("With a request to a deprecated path with an unbounded current outage", func() {
			req := httptest.NewRequest("GET", "http://example.com/unbounded", nil)
			Convey("When the request is made to the router", func() {
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				resp := w.Result()
				Convey("Then the response should be handled by the underlying handler", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
					body, err := io.ReadAll(resp.Body)
					So(err, ShouldBeNil)
					So(body, ShouldResemble, []byte(unboundedMessage+"\n"))
				})
				Convey("And the deprecation headers should be returned", func() {
					headers := resp.Header
					So(headers.Get("Deprecation"), ShouldNotBeEmpty)
					So(headers.Get("Deprecation"), ShouldEqual, unboundedDeprecation)
					So(headers.Get("Link"), ShouldNotBeEmpty)
					So(headers.Get("Link"), ShouldEqual, "<"+unboundedLink+">; rel=\"sunset\"")
					So(headers.Get("Sunset"), ShouldNotBeEmpty)
					So(headers.Get("Sunset"), ShouldEqual, unboundedSunset)
				})
			})
		})

		Convey("With a request to a deprecated path with future unbounded outage", func() {
			req := httptest.NewRequest("GET", "http://example.com/futureunbounded", nil)
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
					So(headers.Get("Deprecation"), ShouldEqual, futureUnboundedDeprecation)
					So(headers.Get("Link"), ShouldNotBeEmpty)
					So(headers.Get("Link"), ShouldEqual, "<"+futureUnboundedLink+">; rel=\"sunset\"")
					So(headers.Get("Sunset"), ShouldNotBeEmpty)
					So(headers.Get("Sunset"), ShouldEqual, futureUnboundedSunset)
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

func anyToPointer[V any](v V) *V {
	return &v
}
