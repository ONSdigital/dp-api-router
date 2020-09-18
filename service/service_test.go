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

func TestNotProxied(t *testing.T) {

	Convey("Given a healthy api router", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		zebedeeURL, err := url.Parse(cfg.ZebedeeURL)
		So(err, ShouldBeNil)

		hcMock := &mock.HealthCheckerMock{
			HandlerFunc: func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		}

		resetProxyMocksWithExpectations(map[string]*url.URL{})

		Convey("A request to a not-registered endpoint falls through to the default zebedee handler", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/wrong", hcMock)
			So(w.Code, ShouldEqual, http.StatusNotFound)
			assertOnlyThisURLIsCalled(zebedeeURL)
		})
	})
}

func TestRouterPublicAPIs(t *testing.T) {

	Convey("Given an api router and proxies with all public endpoints available", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcMock := &mock.HealthCheckerMock{
			HandlerFunc: func(w http.ResponseWriter, req *http.Request) {},
		}

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
		searchAPIURL, err := url.Parse(cfg.SearchAPIURL)
		So(err, ShouldBeNil)
		imageAPIURL, err := url.Parse(cfg.ImageAPIURL)
		So(err, ShouldBeNil)

		expectedPublicURLs := map[string]*url.URL{
			"/code-lists": codelistAPIURL,
			"/datasets":   datasetAPIURL,
			"/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations": observationAPIURL,
			"/filters":        filterAPIURL,
			"/filter-outputs": filterAPIURL,
			"/hierarchies":    hierarchyAPIURL,
			"/search":         searchAPIURL,
			"/images":         imageAPIURL,
		}

		resetProxyMocksWithExpectations(expectedPublicURLs)

		Convey("and observation API enabled by configuration", func() {
			cfg.EnableObservationAPI = true

			Convey("A request to code-list path succeeds and is proxied to codeListAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/code-lists", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/code-lists", codelistAPIURL)
			})

			Convey("A request to code-list subpath succeeds and is proxied to codeListAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/code-lists/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/code-lists/subpath", codelistAPIURL)
			})

			Convey("A request to dataset path succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets", datasetAPIURL)
			})

			Convey("A request to dataset edition path succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123", datasetAPIURL)
			})

			Convey("A request to a dataset edition version endpoint succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123/versions/321", datasetAPIURL)
			})

			Convey("A request to a dataset edition version observation endpoint succeeds and is proxied to observationAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321/observations", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123/versions/321/observations", observationAPIURL)
			})

			Convey("A request to a filters is proxied to filterAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/filters", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/filters", filterAPIURL)
			})

			Convey("A request to a filters subpath is proxied to filterAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/filters/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/filters/subpath", filterAPIURL)
			})

			Convey("A request to a filter-output is proxied to filterAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/filter-outputs", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/filter-outputs", filterAPIURL)
			})

			Convey("A request to a filter-output subpath is proxied to filterAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/filter-outputs/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/filter-outputs/subpath", filterAPIURL)
			})

			Convey("A request to a hierarchies path is proxied to hierarchyAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/hierarchies", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/hierarchies", hierarchyAPIURL)
			})

			Convey("A request to a hierarchies subpath is proxied to hierarchyAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/hierarchies/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/hierarchies/subpath", hierarchyAPIURL)
			})

			Convey("A request to a search path is proxied to searchAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/search", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/search", searchAPIURL)
			})

			Convey("A request to a search subpath is proxied to searchAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/search/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/search/subpath", searchAPIURL)
			})

			Convey("A request to an image subpath is proxied to imageAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/images/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/images/subpath", imageAPIURL)
			})
		})

		Convey("and observation API disabled by configuration", func() {
			cfg.EnableObservationAPI = false

			Convey("A request to dataset path succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets", datasetAPIURL)
			})

			Convey("A request to a dataset edition version observation endpoint succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321/observations", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123/versions/321/observations", datasetAPIURL)
			})

		})

	})
}

func TestRouterPrivateAPIs(t *testing.T) {

	Convey("Given an api router and proxies with all private endpoints available", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcMock := &mock.HealthCheckerMock{}

		datasetAPIURL, err := url.Parse(cfg.DatasetAPIURL)
		So(err, ShouldBeNil)
		recipeAPIURL, err := url.Parse(cfg.RecipeAPIURL)
		So(err, ShouldBeNil)
		importAPIURL, err := url.Parse(cfg.ImportAPIURL)
		So(err, ShouldBeNil)
		zebedeeURL, err := url.Parse(cfg.ZebedeeURL)
		So(err, ShouldBeNil)

		expectedPrivateURLs := map[string]*url.URL{
			"/recipes":   recipeAPIURL,
			"/jobs":      importAPIURL,
			"/instances": datasetAPIURL,
		}

		resetProxyMocksWithExpectations(expectedPrivateURLs)

		Convey("and private endpoints enabled by configuration", func() {
			cfg.EnableObservationAPI = true

			Convey("A request to a recipes path is proxied to recipeAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/recipes", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/recipes", recipeAPIURL)
			})

			Convey("A request to a recipes subpath is proxied to recipeAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/recipes/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/recipes/subpath", recipeAPIURL)
			})

			Convey("A request to a jobs path is proxied to importAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/jobs", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/jobs", importAPIURL)
			})

			Convey("A request to a jobs subpath is proxied to importAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/jobs/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/jobs/subpath", importAPIURL)
			})

			Convey("A request to a instances path is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/instances", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/instances", datasetAPIURL)
			})

			Convey("A request to a instances subpath is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/instances/subpath", hcMock)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/instances/subpath", datasetAPIURL)
			})
		})

		Convey("and private endpoints disabled by configuration", func() {
			cfg.EnablePrivateEndpoints = false

			Convey("A request to a recipes path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/recipes", hcMock)
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to a jobs path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/jobs", hcMock)
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to an instances path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/instances", hcMock)
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

		})
	})
}

func TestRouterLegacyAPIs(t *testing.T) {

	Convey("Given an api router and proxies with all legacy endpoints available", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcMock := &mock.HealthCheckerMock{}

		apiPocURL, err := url.Parse(cfg.APIPocURL)
		So(err, ShouldBeNil)

		expectedLegacyURLs := map[string]*url.URL{
			"/ops":        apiPocURL,
			"/dataset":    apiPocURL,
			"/timeseries": apiPocURL,
			"/search":     apiPocURL,
		}

		resetProxyMocksWithExpectations(expectedLegacyURLs)

		Convey("An un-versioned request to an ops path is proxied to pocAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/ops", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/ops", apiPocURL)
		})

		Convey("An un-versioned request to a dataset path is proxied to pocAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/dataset", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/dataset", apiPocURL)
		})

		Convey("An un-versioned request to a timeseries path is proxied to pocAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/timeseries", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/timeseries", apiPocURL)
		})

		Convey("An un-versioned request to a search path is proxied to pocAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/search", hcMock)
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/search", apiPocURL)
		})
	})
}

func assertOnlyThisURLIsCalled(expectedURL *url.URL) {
	for url, pxy := range registeredProxies {

		if url == *expectedURL {
			So(len(pxy.ServeHTTPCalls()), ShouldEqual, 1)
			continue
		}

		So(len(pxy.ServeHTTPCalls()), ShouldEqual, 0)
	}
}

// createRouterTest calls service CreateRouter httptest request, recorder, and healthcheck mock
func createRouterTest(cfg *config.Config, url string, hcMock *mock.HealthCheckerMock) *httptest.ResponseRecorder {
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

// resetProxyMocksWithExpectations resets the global variable `registeredProxies`, and sets the NewSingleHostReverseProxy to return
// a proxy mock according to the expected URLs map.
func resetProxyMocksWithExpectations(expectedURLs map[string]*url.URL) {
	registeredProxies = map[url.URL]*proxyMock.IReverseProxyMock{}

	proxy.NewSingleHostReverseProxy = func(target *url.URL, version, envHost, contextURL string) proxy.IReverseProxy {
		pxyMock := &proxyMock.IReverseProxyMock{
			ServeHTTPFunc: func(rw http.ResponseWriter, req *http.Request) {

				for path := range expectedURLs {
					if strings.HasPrefix(req.URL.Path, path) {
						return
					}
				}
				http.Error(rw, "path not found", http.StatusNotFound)
			},
		}
		registeredProxies[*target] = pxyMock
		return pxyMock
	}
}
