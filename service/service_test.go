package service_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-api-router/proxy"
	proxyMock "github.com/ONSdigital/dp-api-router/proxy/mock"
	"github.com/ONSdigital/dp-api-router/service"

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

		resetProxyMocksWithExpectations(map[string]*url.URL{})

		Convey("A request to a not-registered endpoint falls through to the default zebedee handler", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/wrong")
			So(w.Code, ShouldEqual, http.StatusNotFound)
			assertOnlyThisURLIsCalled(zebedeeURL)
		})
	})
}

func TestRouterPublicAPIs(t *testing.T) {
	Convey("Given an api router and proxies with all public endpoints available", t, func() {
		cfg, _ := config.Get()

		zebedeeURL, _ := url.Parse(cfg.ZebedeeURL)
		hierarchyAPIURL, _ := url.Parse(cfg.HierarchyAPIURL)
		filterAPIURL, _ := url.Parse(cfg.FilterAPIURL)
		filterFlexAPIURL, _ := url.Parse(cfg.FilterFlexAPIURL)
		datasetAPIURL, _ := url.Parse(cfg.DatasetAPIURL)
		observationAPIURL, _ := url.Parse(cfg.ObservationAPIURL)
		codelistAPIURL, _ := url.Parse(cfg.CodelistAPIURL)
		searchAPIURL, _ := url.Parse(cfg.SearchAPIURL)
		dimensionSearchAPIURL, _ := url.Parse(cfg.DimensionSearchAPIURL)
		imageAPIURL, _ := url.Parse(cfg.ImageAPIURL)
		feedbackAPIURL, _ := url.Parse(cfg.FeedbackAPIURL)
		releaseCalendarAPIURL, _ := url.Parse(cfg.ReleaseCalendarAPIURL)
		populationTypesAPIURL, _ := url.Parse(cfg.PopulationTypesAPIURL)
		topicAPIURL, _ := url.Parse(cfg.TopicAPIURL)
		scrubberAPIURL, _ := url.Parse(cfg.SearchScrubberAPIURL)
		categoryAPIURL, _ := url.Parse(cfg.CategoryAPIURL)
		berlinAPIURL, _ := url.Parse(cfg.BerlinAPIURL)

		expectedPublicURLs := map[string]*url.URL{
			"/code-lists": codelistAPIURL,
			"/datasets":   datasetAPIURL,
			"/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations":        observationAPIURL,
			"/datasets/{dataset_id}/editions/{edition}/versions/{version}/json":                filterFlexAPIURL,
			"/datasets/{dataset_id}/editions/{edition}/versions/{version}/census-observations": filterFlexAPIURL,
			"/custom/filters":   filterFlexAPIURL,
			"/filters":          filterAPIURL,
			"/filter-outputs":   filterAPIURL,
			"/hierarchies":      hierarchyAPIURL,
			"/search":           searchAPIURL,
			"/dimension-search": dimensionSearchAPIURL,
			"/images":           imageAPIURL,
			"/population-types": populationTypesAPIURL,
			"/navigation":       topicAPIURL,
			"/topics":           topicAPIURL,
			"/berlin":           berlinAPIURL,
			"/scrubber":         scrubberAPIURL,
			"/categories":       categoryAPIURL,
		}

		cfg.FeedbackAPIVersions = []string{"aB", "bE"}
		for _, version := range cfg.FeedbackAPIVersions {
			expectedPublicURLs["/"+version+"/feedback"] = feedbackAPIURL
		}
		cfg.ReleaseCalendarAPIVersions = []string{"vX", "vY"}
		for _, version := range cfg.ReleaseCalendarAPIVersions {
			expectedPublicURLs["/"+version+"/releases"] = releaseCalendarAPIURL
		}

		resetProxyMocksWithExpectations(expectedPublicURLs)

		Convey("A request to code-list path succeeds and is proxied to codeListAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/code-lists")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/code-lists", codelistAPIURL)
		})

		Convey("A request to code-list subpath succeeds and is proxied to codeListAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/code-lists/subpath")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/code-lists/subpath", codelistAPIURL)
		})

		Convey("When the enable observation feature flag is enabled", func() {
			cfg.EnableObservationAPI = true

			Convey("A request to dataset path succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets", datasetAPIURL)
			})

			Convey("A request to dataset edition path succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123", datasetAPIURL)
			})

			Convey("A request to a dataset edition version endpoint succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123/versions/321", datasetAPIURL)
			})

			Convey("A request to a dataset edition version observation endpoint succeeds and is proxied to observationAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321/observations")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123/versions/321/observations", observationAPIURL)
			})
		})

		Convey("When the enable observation feature flag is disabled", func() {
			cfg.EnableObservationAPI = false

			Convey("A request to dataset path succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets", datasetAPIURL)
			})

			Convey("A request to dataset edition path succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123", datasetAPIURL)
			})

			Convey("A request to a dataset edition version endpoint succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123/versions/321", datasetAPIURL)
			})

			Convey("A request to a dataset edition version observation endpoint succeeds and is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/cpih012/editions/123/versions/321/observations")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/datasets/cpih012/editions/123/versions/321/observations", datasetAPIURL)
			})
		})

		Convey("A request to a dataset edition version census observation endpoint succeeds and is proxied to filterFlexAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/datasets/RM111/editions/123/versions/321/census-observations")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/datasets/RM111/editions/123/versions/321/census-observations", filterFlexAPIURL)
		})

		Convey("A request to a filters is proxied to filterAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/filters")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/filters", filterAPIURL)
		})

		Convey("A request to a filters subpath is proxied to filterAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/filters/subpath")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/filters/subpath", filterAPIURL)
		})

		Convey("A request to a filter-output is proxied to filterAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/filter-outputs")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/filter-outputs", filterAPIURL)
		})

		Convey("A request to a filter-output subpath is proxied to filterAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/filter-outputs/subpath")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/filter-outputs/subpath", filterAPIURL)
		})

		Convey("A request to a hierarchies path is proxied to hierarchyAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/hierarchies")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/hierarchies", hierarchyAPIURL)
		})

		Convey("A request to a hierarchies subpath is proxied to hierarchyAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/hierarchies/subpath")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/hierarchies/subpath", hierarchyAPIURL)
		})

		Convey("A request to a search path is proxied to searchAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/search")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/search", searchAPIURL)
		})

		Convey("A request to a search subpath is proxied to searchAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/search/subpath")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/search/subpath", searchAPIURL)
		})

		Convey("A request to a dimension search path is proxied to dimensionSearchAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/dimension-search")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/dimension-search", dimensionSearchAPIURL)
		})

		Convey("A request to a dimension search subpath is proxied to dimensionSearchAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/dimension-search/subpath")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/dimension-search/subpath", dimensionSearchAPIURL)
		})

		Convey("A request to an image subpath is proxied to imageAPIURL", func() {
			w := createRouterTest(cfg, "http://localhost:23200/v1/images/subpath")
			So(w.Code, ShouldEqual, http.StatusOK)
			verifyProxied("/images/subpath", imageAPIURL)
		})

		Convey("Given an URL to an feedback subpath", func() {
			host := "http://localhost:23200"
			path := "/%s/feedback/subpath"
			urlPattern := host + path

			Convey("And the feature flag is enabled", func() {
				cfg.EnableFeedbackAPI = true
				for _, version := range cfg.FeedbackAPIVersions {
					Convey("When we make a GET request using the mapped version "+version, func() {
						w := createRouterTest(cfg, fmt.Sprintf(urlPattern, version))
						Convey("Then the request is proxied to the feedback API", func() {
							So(w.Code, ShouldEqual, http.StatusOK)
							verifyProxied(fmt.Sprintf(path, version), feedbackAPIURL)
						})
					})
				}

				Convey("When we make a GET request using an unmapped version", func() {
					version := "v9"
					createRouterTest(cfg, fmt.Sprintf(urlPattern, version))
					Convey("Then the request falls through to the default zebedee handler", func() {
						verifyProxied(fmt.Sprintf(path, version), zebedeeURL)
					})
				})
			})

			Convey("And the feature flag is disabled", func() {
				cfg.EnableFeedbackAPI = false
				for _, version := range cfg.FeedbackAPIVersions {
					Convey("When we make a GET request using the mapped version "+version, func() {
						// deliberately not configured v1 to get around legacy handle stripping it
						w := createRouterTest(cfg, fmt.Sprintf(urlPattern, version))
						Convey("Then it falls through to the default zebedee handler", func() {
							So(w.Code, ShouldEqual, http.StatusOK)
							verifyProxied(fmt.Sprintf(path, version), zebedeeURL)
						})
					})
				}
			})
		})

		Convey("A request to a release calendar subpath", func() {
			host := "http://localhost:23200"
			path := "/%s/releases/subpath"
			urlPattern := host + path

			Convey("When the feature flag is enabled", func() {
				cfg.EnableReleaseCalendarAPI = true

				for _, version := range cfg.ReleaseCalendarAPIVersions {
					Convey("And the request is using the mapped version "+version, func() {
						Convey("Then the request is proxied to the release calendar API", func() {
							w := createRouterTest(cfg, fmt.Sprintf(urlPattern, version))
							So(w.Code, ShouldEqual, http.StatusOK)
							verifyProxied(fmt.Sprintf(path, version), releaseCalendarAPIURL)
						})
					})
				}

				Convey("And the request is using an unmapped version", func() {
					Convey("Then the request falls through to the default zebedee handler", func() {
						version := "v9"
						createRouterTest(cfg, fmt.Sprintf(urlPattern, version))
						verifyProxied(fmt.Sprintf(path, version), zebedeeURL)
					})
				})
			})

			Convey("When the feature flag is disabled", func() {
				cfg.EnableReleaseCalendarAPI = false
				Convey("Then all requests falls through to the default zebedee handler", func() {
					for _, version := range cfg.ReleaseCalendarAPIVersions {
						// deliberately not configured v1 to get around legacy handle stripping it
						w := createRouterTest(cfg, fmt.Sprintf(urlPattern, version))
						So(w.Code, ShouldEqual, http.StatusOK)
						verifyProxied(fmt.Sprintf(path, version), zebedeeURL)
					}
				})
			})
		})

		Convey("Given an url to the population types api", func() {
			urlStr := "http://localhost:23200/v1/population-types"

			Convey("And the feature flag is enabled", func() {
				cfg.EnablePopulationTypesAPI = true
				Convey("When a GET request is made", func() {
					w := createRouterTest(cfg, urlStr)
					Convey("Then the population types API should respond", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
						verifyProxied("/population-types", populationTypesAPIURL)
					})
				})
			})

			Convey("And the feature flag is disabled", func() {
				cfg.EnablePopulationTypesAPI = false
				Convey("When a GET request is made", func() {
					w := createRouterTest(cfg, urlStr)
					Convey("Then the default zebedee handler should respond", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
						verifyProxied("/population-types", zebedeeURL)
					})
				})
			})
		})

		Convey("Given a topic service path", func() {
			topicsPath := "http://localhost:23200/v1/topics"

			Convey("Then a request to the topics endpoint is proxied to the topicsAPIURL", func() {
				w := createRouterTest(cfg, topicsPath)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/topics", topicAPIURL)
			})
		})

		Convey("Given a navigation path", func() {
			navigationPath := "http://localhost:23200/v1/navigation"

			Convey("Then a request to the navigation endpoint is proxied to topicsAPIURL", func() {
				w := createRouterTest(cfg, navigationPath)
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/navigation", topicAPIURL)
			})
		})

		Convey("When the enable NLP APIs flag is enabled", func() {
			cfg.EnableNLPSearchAPIs = true

			Convey("Then a request to the scrubber endpoint is proxied to the scrubber URLS", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/scrubber")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/scrubber", scrubberAPIURL)
			})

			Convey("Then a request to the berlin endpoint is proxied to the berlin URLS", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/berlin")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/berlin", berlinAPIURL)
			})

			Convey("Then a request to the category endpoint is proxied to the category URLS", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/categories")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/categories", categoryAPIURL)
			})
		})

		Convey("When the enable NLP APIs flag is disabled", func() {
			cfg.EnableNLPSearchAPIs = false

			Convey("Then a request to the scrubber endpoint is proxied to the Zebedee", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/scrubber")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/scrubber", zebedeeURL)
			})

			Convey("Then a request to the berlin endpoint is proxied to the Zebedee", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/berlin")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/berlin", zebedeeURL)
			})

			Convey("Then a request to the category endpoint is proxied to Zebedee", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/categories")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/categories", zebedeeURL)
			})
		})
	})
}

func TestRouterPrivateAPIs(t *testing.T) {
	Convey("Given an api router and proxies with all private endpoints available", t, func() {
		cfg, _ := config.Get()

		datasetAPIURL, _ := url.Parse(cfg.DatasetAPIURL)
		recipeAPIURL, _ := url.Parse(cfg.RecipeAPIURL)
		importAPIURL, _ := url.Parse(cfg.ImportAPIURL)
		uploadServiceAPIURL, _ := url.Parse(cfg.UploadServiceAPIURL)
		identityAPIURL, _ := url.Parse(cfg.IdentityAPIURL)
		permissionsAPIURL, _ := url.Parse(cfg.PermissionsAPIURL)
		zebedeeURL, _ := url.Parse(cfg.ZebedeeURL)
		cantabularMetadataExtractorAPIURL, _ := url.Parse(cfg.CantabularMetadataExtractorAPIURL)
		redirectAPIURL, _ := url.Parse(cfg.RedirectAPIURL)
		bundleAPIURL, _ := url.Parse(cfg.BundleAPIURL)

		expectedPrivateURLs := map[string]*url.URL{
			"/upload":              uploadServiceAPIURL,
			"/recipes":             recipeAPIURL,
			"/jobs":                importAPIURL,
			"/instances":           datasetAPIURL,
			"/cantabular-metadata": cantabularMetadataExtractorAPIURL,
			"/redirects":           redirectAPIURL,
			"/bundles":             bundleAPIURL,
			"/bundle-events":       bundleAPIURL,
		}
		for _, version := range cfg.IdentityAPIVersions {
			expectedPrivateURLs[fmt.Sprintf("/%s/tokens", version)] = identityAPIURL
			expectedPrivateURLs[fmt.Sprintf("/%s/users", version)] = identityAPIURL
			expectedPrivateURLs[fmt.Sprintf("/%s/groups", version)] = identityAPIURL
			expectedPrivateURLs[fmt.Sprintf("/%s/password-reset", version)] = identityAPIURL
		}
		for _, version := range cfg.PermissionsAPIVersions {
			expectedPrivateURLs[fmt.Sprintf("/%s/policies", version)] = permissionsAPIURL
			expectedPrivateURLs[fmt.Sprintf("/%s/roles", version)] = permissionsAPIURL
			expectedPrivateURLs[fmt.Sprintf("/%s/permissions-bundle", version)] = permissionsAPIURL
		}

		resetProxyMocksWithExpectations(expectedPrivateURLs)

		Convey("and private endpoints enabled by configuration", func() {
			cfg.EnablePrivateEndpoints = true

			Convey("A request to a recipes path is proxied to recipeAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/recipes")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/recipes", recipeAPIURL)
			})

			Convey("A request to a recipes subpath is proxied to recipeAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/recipes/subpath")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/recipes/subpath", recipeAPIURL)
			})

			Convey("A request to a jobs path is proxied to importAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/jobs")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/jobs", importAPIURL)
			})

			Convey("A request to a jobs subpath is proxied to importAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/jobs/subpath")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/jobs/subpath", importAPIURL)
			})

			Convey("A request to a jobs path is proxied to uploadServiceAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/upload")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/upload", uploadServiceAPIURL)
			})

			Convey("A request to a jobs subpath is proxied to uploadServiceAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/upload/subpath")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/upload/subpath", uploadServiceAPIURL)
			})

			Convey("A request to a instances path is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/instances")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/instances", datasetAPIURL)
			})

			Convey("A request to a instances subpath is proxied to datasetAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/instances/subpath")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/instances/subpath", datasetAPIURL)
			})

			Convey("A request to a tokens path is proxied to identityAPIURL", func() {
				for _, version := range cfg.IdentityAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/tokens")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/tokens", identityAPIURL)
				}
			})

			Convey("A request to a tokens subpath is proxied to identityAPIURL", func() {
				for _, version := range cfg.IdentityAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/tokens/subpath")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/tokens/subpath", identityAPIURL)
				}
			})

			Convey("A request to a users path is proxied to identityAPIURL", func() {
				for _, version := range cfg.IdentityAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/users")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/users", identityAPIURL)
				}
			})

			Convey("A request to a users subpath is proxied to identityAPIURL", func() {
				for _, version := range cfg.IdentityAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/users/subpath")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/users/subpath", identityAPIURL)
				}
			})

			Convey("A request to a groups path is proxied to identityAPIURL", func() {
				for _, version := range cfg.IdentityAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/groups")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/groups", identityAPIURL)
				}
			})

			Convey("A request to a groups subpath is proxied to identityAPIURL", func() {
				for _, version := range cfg.IdentityAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/groups/subpath")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/groups/subpath", identityAPIURL)
				}
			})

			Convey("A request to password-reset path is proxied to identityAPIURL", func() {
				for _, version := range cfg.IdentityAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/password-reset")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/password-reset", identityAPIURL)
				}
			})

			Convey("A request to policies path is proxied to permissionsAPIURL", func() {
				for _, version := range cfg.PermissionsAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/policies")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/policies", permissionsAPIURL)
				}
			})

			Convey("A request to roles path is proxied to permissionsAPIURL", func() {
				for _, version := range cfg.PermissionsAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/roles")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/roles", permissionsAPIURL)
				}
			})

			Convey("A request to permissions-bundle path is proxied to permissionsAPIURL", func() {
				for _, version := range cfg.PermissionsAPIVersions {
					w := createRouterTest(cfg, "http://localhost:23200/"+version+"/permissions-bundle")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/"+version+"/permissions-bundle", permissionsAPIURL)
				}
			})

			Convey("When the enable cantabular metadata extractor API feature flag is enabled", func() {
				cfg.EnableCantabularMetadataExtractorAPI = true

				Convey("A request to a cantabular-metadata subpath is proxied to cantabularMetadataExtractorAPIURL", func() {
					w := createRouterTest(cfg, "http://localhost:23200/v1/cantabular-metadata/subpath")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/cantabular-metadata/subpath", cantabularMetadataExtractorAPIURL)
				})
			})

			Convey("When the enable cantabular metadata extractor API feature flag is disabled", func() {
				cfg.EnableCantabularMetadataExtractorAPI = false

				Convey("A request to a cantabular-metadata subpath is not proxied", func() {
					createRouterTest(cfg, "http://localhost:23200/v1/cantabular-metadata/subpath")
					assertOnlyThisURLIsCalled(zebedeeURL)
				})
			})

			Convey("When the redirect api feature flag is enabled", func() {
				cfg.EnableRedirectAPI = true

				Convey("Then a request to the redirects path is proxied to the redirectAPIURL", func() {
					w := createRouterTest(cfg, "http://localhost:23200/v1/redirects")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/redirects", redirectAPIURL)
				})

				Convey("Then a request to the redirects subpath is proxied to the redirectAPIURL", func() {
					w := createRouterTest(cfg, "http://localhost:23200/v1/redirects/subpath")
					So(w.Code, ShouldEqual, http.StatusOK)
					verifyProxied("/redirects/subpath", redirectAPIURL)
				})
			})

			Convey("When the redirect api feature flag is disabled", func() {
				cfg.EnableRedirectAPI = false

				Convey("Then a request to the redirects path is proxied to Zebedee", func() {
					createRouterTest(cfg, "http://localhost:23200/v1/redirects")
					assertOnlyThisURLIsCalled(zebedeeURL)
				})

				Convey("Then a request to the redirects subpath is proxied to Zebedee", func() {
					createRouterTest(cfg, "http://localhost:23200/v1/redirects/subpath")
					assertOnlyThisURLIsCalled(zebedeeURL)
				})
			})
		})

		Convey("When the bundle API feature flag is enabled", func() {
			cfg.EnableBundleAPI = true

			Convey("A request to a bundles subpath is proxied to bundleAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/bundles/subpath")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/bundles/subpath", bundleAPIURL)
			})

			Convey("A request to a bundle-events subpath is proxied to bundleAPIURL", func() {
				w := createRouterTest(cfg, "http://localhost:23200/v1/bundle-events/subpath")
				So(w.Code, ShouldEqual, http.StatusOK)
				verifyProxied("/bundle-events/subpath", bundleAPIURL)
			})
		})

		Convey("When the bundle API feature flag is disabled", func() {
			cfg.EnableBundleAPI = false

			Convey("A request to a bundles subpath is not proxied", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/bundles/subpath")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})
		})

		Convey("and private endpoints disabled by configuration", func() {
			cfg.EnablePrivateEndpoints = false

			Convey("A request to a recipes path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/recipes")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to a jobs path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/jobs")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to an instances path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/instances")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to a policies path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/policies")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to a roles path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/roles")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to a permissions-bundle path is not proxied and fails with StatusNotFound", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/permissions-bundle")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})

			Convey("A request to a cantabular-metadata subpath is not proxied", func() {
				createRouterTest(cfg, "http://localhost:23200/v1/cantabular-metadata/subpath")
				assertOnlyThisURLIsCalled(zebedeeURL)
			})
		})
	})
}

func assertOnlyThisURLIsCalled(expectedURL *url.URL) {
	for urlToCheck, pxy := range registeredProxies {
		if urlToCheck == *expectedURL {
			So(len(pxy.ServeHTTPCalls()), ShouldEqual, 1)
			continue
		}
		So(len(pxy.ServeHTTPCalls()), ShouldEqual, 0)
	}
}

// createRouterTest calls service CreateRouter httptest request, recorder, and healthcheck mock
func createRouterTest(cfg *config.Config, urlStr string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, urlStr, http.NoBody)
	r.Header.Set(authorizationHeader, testServiceAuthToken)
	w := httptest.NewRecorder()

	router := service.CreateRouter(testCtx, cfg)
	router.ServeHTTP(w, r)
	return w
}

// verifyProxied asserts that only the proxy that was registered for the expected URL is called, with the expected path
func verifyProxied(path string, expectedURL *url.URL) {
	pxy, found := registeredProxies[*expectedURL]
	So(found, ShouldBeTrue)
	So(len(pxy.ServeHTTPCalls()), ShouldEqual, 1)
	So(pxy.ServeHTTPCalls()[0].Req.Header.Get(authorizationHeader), ShouldEqual, testServiceAuthToken)
	So(pxy.ServeHTTPCalls()[0].Req.URL.Path, ShouldEqual, path)
	for urlToCheck, pxy := range registeredProxies {
		if urlToCheck != *expectedURL {
			So(len(pxy.ServeHTTPCalls()), ShouldEqual, 0)
		}
	}
}

// resetProxyMocksWithExpectations resets the global variable `registeredProxies`, and sets the NewSingleHostReverseProxy to return
// a proxy mock according to the expected URLs map.
func resetProxyMocksWithExpectations(expectedURLs map[string]*url.URL) {
	registeredProxies = map[url.URL]*proxyMock.IReverseProxyMock{}

	proxy.NewSingleHostReverseProxyWithTransport = func(target *url.URL, transport http.RoundTripper) proxy.IReverseProxy {
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
