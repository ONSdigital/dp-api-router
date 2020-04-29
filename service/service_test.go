package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-api-router/proxy"
	proxyMock "github.com/ONSdigital/dp-api-router/proxy/mock"
	"github.com/ONSdigital/dp-api-router/service"
	"github.com/ONSdigital/dp-api-router/service/mock"
	. "github.com/smartystreets/goconvey/convey"
)

const authorizationHeader = "Authorization"

var (
	testCtx              = context.Background()
	registeredProxies    = map[url.URL]*proxyMock.IReverseProxyMock{}
	testServiceAuthToken = "Bearer testServiceAuthToken"
)

func TestRouterPublicAPIs(t *testing.T) {

	Convey("Given a healthy api router and a proxy with all expected endpoints available", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcMock := &mock.IHealthCheckMock{
			HandlerFunc: func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		}

		// Parse URL to set expectations
		hierarchyAPIURL, err := url.Parse(cfg.HierarchyAPIURL)
		So(err, ShouldBeNil)
		filterAPIURL, err := url.Parse(cfg.FilterAPIURL)
		So(err, ShouldBeNil)
		datasetAPIURL, err := url.Parse(cfg.DatasetAPIURL)
		So(err, ShouldBeNil)
		observationAPIURL, err := url.Parse(cfg.ObservationAPIURL)
		So(err, ShouldBeNil)
		codelistAPIURL, err := url.Parse(cfg.CodelistAPIURL)
		So(err, ShouldBeNil)
		recipeAPIURL, err := url.Parse(cfg.RecipeAPIURL)
		So(err, ShouldBeNil)
		importAPIURL, err := url.Parse(cfg.ImportAPIURL)
		So(err, ShouldBeNil)
		searchAPIURL, err := url.Parse(cfg.SearchAPIURL)
		So(err, ShouldBeNil)
		aPIPocURL, err := url.Parse(cfg.APIPocURL)
		So(err, ShouldBeNil)

		expectedPublicURLs := map[string]*url.URL{
			"/code-lists": codelistAPIURL,
			"/datasets":   datasetAPIURL,
			"/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations": observationAPIURL,
			"/filters":        filterAPIURL,
			"/filter-outputs": filterAPIURL,
			"/hierarchies":    hierarchyAPIURL,
			"/search":         searchAPIURL,
			// }

			// expectedPrivateURLs := map[string]*url.URL{
			"/recipes":   recipeAPIURL,
			"/jobs":      importAPIURL,
			"/instances": datasetAPIURL,
			// }

			// legacyURLs := map[string]*url.URL{
			"/ops":        aPIPocURL,
			"/dataset":    aPIPocURL,
			"/timeseries": aPIPocURL,
			// "/search":     aPIPocURL,
		}

		registeredProxies = map[url.URL]*proxyMock.IReverseProxyMock{}

		proxy.NewSingleHostReverseProxy = func(target *url.URL, version, envHost, contextURL string) proxy.IReverseProxy {
			pxyMock := &proxyMock.IReverseProxyMock{
				ServeHTTPFunc: func(rw http.ResponseWriter, req *http.Request) {
					for path := range expectedPublicURLs {
						if strings.HasPrefix(req.URL.Path, path) {
							return
						}
					}
					http.Error(rw, "wrong path", http.StatusBadGateway)
				},
			}
			registeredProxies[*target] = pxyMock
			return pxyMock
		}

		Convey("A request to the health endpoint is successful and not proxied", func() {
			w := createRouterTest("http://localhost:23200/health", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			for _, pxy := range registeredProxies {
				So(len(pxy.ServeHTTPCalls()), ShouldEqual, 0)
			}
		})

		Convey("A request to a wrong endpoint fails with Status NotFound and is not proxied", func() {
			w := createRouterTest("http://localhost:23200/v1/wrong", hcMock)
			So(w.Code, ShouldEqual, http.StatusNotFound)
			for _, pxy := range registeredProxies {
				So(len(pxy.ServeHTTPCalls()), ShouldEqual, 0)
			}
		})

		Convey("A request to code-list path succeeds and is proxied to codeListAPIURL", func() {
			w := createRouterTest("http://localhost:23200/v1/code-lists", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/code-lists", codelistAPIURL)
		})

		Convey("A request to code-list subpath succeeds and is proxied to codeListAPIURL", func() {
			w := createRouterTest("http://localhost:23200/v1/code-lists/subpath", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/code-lists/subpath", codelistAPIURL)
		})

		Convey("A request to dataset path succeeds and is proxied to datasetAPIURL", func() {
			w := createRouterTest("http://localhost:23200/v1/datasets", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/datasets", datasetAPIURL)
		})

		Convey("A request to dataset edition path succeeds and is proxied to datasetAPIURL", func() {
			w := createRouterTest("http://localhost:23200/v1/datasets/cpih012/editions/123", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/datasets/cpih012/editions/123", datasetAPIURL)
		})

		Convey("A request to a dataset edition version endpoint succeeds and is proxied to datasetAPIURL", func() {
			w := createRouterTest("http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/datasets/cpih012/editions/123/versions/321", datasetAPIURL)
		})

		Convey("A request to a dataset edition version endpoint", func() {
			w := createRouterTest("http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321/observations", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/datasets/cpih012/editions/123/versions/321/observations", observationAPIURL)
		})

	})
}

// createRouterTest calls service CreateRouter httptest request, recorder, and healthcheck mock
func createRouterTest(url string, hcMock *mock.IHealthCheckMock) *httptest.ResponseRecorder {
	cfg, err := config.Get()
	So(err, ShouldBeNil)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r.Header.Set(authorizationHeader, testServiceAuthToken)
	w := httptest.NewRecorder()
	router := service.CreateRouter(testCtx, cfg, hcMock)
	router.ServeHTTP(w, r)
	return w
}

// verifyProxied asserts tha only the proxy that was registered for the expected URL is called, with the expected path
func verifyProxied(path string, expectedURL *url.URL) {
	pxy, found := registeredProxies[*expectedURL]
	So(found, ShouldBeTrue)
	So(len(pxy.ServeHTTPCalls()), ShouldEqual, 1)
	So(pxy.ServeHTTPCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, testServiceAuthToken)
	So(pxy.ServeHTTPCalls()[0].Req.URL.Path, ShouldEqual, path)
	for url, pxy := range registeredProxies {
		if url != *expectedURL {
			So(len(pxy.ServeHTTPCalls()), ShouldEqual, 0)
		}
	}
}
