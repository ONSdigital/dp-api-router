package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	"github.com/ONSdigital/dp-api-router/event"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out ./mock/router.go -pkg mock . Router

type Router interface {
	Match(req *http.Request, match *mux.RouteMatch) bool
}

// paths that will skip auditing (note)
// identity api paths being added until the auditing has been updated to work with new tokens
var pathsToIgnore = []string{
	"/ping",
	"/clickEventLog",
	"/health",
}

// paths that will skip retrieveIdentity, and will be audited without identity
var pathsSkipIdentity = []string{
	"/login",
	"/password",
	"/hierarchies",
	"/tokens",
	"/password-reset",
	"/users/self/password",
}

// Now is a time.Now wrapper specifically for testing purposes, and should not me unlambda'd - despite what golangci-lint says
var Now = time.Now

func ShallSkipIdentity(versionPrefix, path string) bool {
	// TODO need to revisit this if we start supporting multiple versions of the APIs.
	path = strings.TrimPrefix(path, fmt.Sprintf("/%s", versionPrefix))

	for _, pathSkipIdentity := range pathsSkipIdentity {
		if strings.HasPrefix(path, pathSkipIdentity) {
			return true
		}
	}
	return false
}

func shallIgnore(path string) bool {
	for _, pathToIgnore := range pathsToIgnore {
		if strings.HasPrefix(path, pathToIgnore) {
			return true
		}
	}
	return false
}

// AuditHandler is a middleware handler that keeps track of calls for auditing purposes,
// before and after proxying calling the downstream service.
// It obtains the user and caller information by calling Zebedee GET /identity
func AuditHandler(auditProducer *event.AvroProducer,
	cli dphttp.Clienter,
	zebedeeURL, versionPrefix string,
	enableZebedeeAudit bool,
	router Router,
	auth authorisation.Config) func(h http.Handler) http.Handler {
	// create Identity client that will be used by middleware to check callers identity
	idClient := clientsidentity.NewWithHealthClient(&health.Client{
		Client: cli,
		URL:    zebedeeURL,
		Name:   "identity",
	})

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// if path does not need to be audited, ignore it and proceed to next handler
			if shallIgnore(r.URL.Path) {
				h.ServeHTTP(w, r)
				return
			}

			if !enableZebedeeAudit {
				// Zebedee is the fallback route, so we can only determine a request is for Zebedee
				// if it does not match any of the configured routes.
				var matchedRoute = &mux.RouteMatch{}
				router.Match(r, matchedRoute) // bool return value still returns true for NotFoundHandler
				if matchedRoute.MatchErr == mux.ErrNotFound {
					h.ServeHTTP(w, r)
					return
				}
			}

			// Inbound audit event (before proxying).
			auditEvent := GenerateAuditEvent(r)

			if !ShallSkipIdentity(versionPrefix, r.URL.Path) {
				// Retrieve Identity from Zebedee, which is stored in context.
				// if it fails, try to audit with the statusCode before returning
				ctx, statusCode, err := retrieveIdentity(w, r, idClient, auth)
				if err != nil {
					// error already handled in retrieveIdentity. Try to audit it.
					auditEvent.StatusCode = int32(statusCode)
					if auditErr := auditProducer.Audit(auditEvent); auditErr != nil {
						log.Error(ctx, "inbound audit event could not be sent", auditErr, log.Data{"event": auditEvent})
					}
					return
				}
				r = r.WithContext(ctx)

				// Add identity to audit event. User identity takes priority over service identity.
				// If no identity is available, then try to audit without identity and fail the request.
				userIdentity := dprequest.User(ctx)
				serviceIdentity := dprequest.Caller(ctx)
				if userIdentity != "" {
					auditEvent.Identity = userIdentity
				} else if serviceIdentity != "" {
					auditEvent.Identity = serviceIdentity
				} else {
					handleError(ctx, w, r, http.StatusUnauthorized, "", err, log.Data{"event": auditEvent})
					auditEvent.StatusCode = int32(http.StatusUnauthorized)
					if err := auditProducer.Audit(auditEvent); err != nil {
						log.Error(ctx, "inbound audit event could not be sent", err, log.Data{"event": auditEvent})
					}
					return
				}
			}

			// Acceptable request. Audit it before proxying.
			if err := auditProducer.Audit(auditEvent); err != nil {
				handleError(r.Context(), w, r, http.StatusInternalServerError, "inbound audit event could not be sent", err, log.Data{"event": auditEvent})
				return
			}

			// Proxy the call with our responseRecorder
			rec := &responseRecorder{w, http.StatusOK, &bytes.Buffer{}}
			h.ServeHTTP(rec, r)

			// Audit event (after proxying).
			auditEvent.CreatedAt = event.CreatedAtMillis(Now())
			auditEvent.StatusCode = int32(rec.statusCode)
			eventBytes, err := auditProducer.Marshal(auditEvent)
			if err != nil {
				handleError(r.Context(), w, r, http.StatusInternalServerError, "outbound audit event could not be sent", err, log.Data{"event": auditEvent})
				return
			}

			// Copy the intercepted body and header to the original response writer
			w.WriteHeader(rec.statusCode)
			_, err = io.Copy(w, rec.body)
			if err != nil {
				handleError(r.Context(), w, r, http.StatusInternalServerError, "failed to copy intercepted response to original response writer", err, log.Data{"event": auditEvent})
				return
			}

			// Finally send the outbound audit message
			auditProducer.Send(eventBytes)
		})
	}
}

// responseRecorder implements ResponseWriter and intercepts status code and body
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// WriteHeader intercepts the status code and stores it in rec.statusCode instead of setting it to the wrapped responseWriter
func (rec *responseRecorder) WriteHeader(status int) {
	rec.statusCode = status
}

// Write intercepts the body and stores it in rec.body instead of writing it to the wrapped responseWriter
func (rec *responseRecorder) Write(b []byte) (int, error) {
	return rec.body.Write(b)
}

// GenerateAuditEvent creates an audit event with the values from request and request context, if present.
func GenerateAuditEvent(req *http.Request) *event.Audit {
	auditEvent := &event.Audit{
		CreatedAt: event.CreatedAtMillis(Now()),
		Path:      req.URL.Path,
		Method:    req.Method,
	}

	// obtain collectionID from context
	auditEvent.RequestID = dprequest.GetRequestId(req.Context())
	if colID := req.Header.Get(dprequest.CollectionIDHeaderKey); colID != "" {
		auditEvent.CollectionID = colID
	}

	// try to unescape query parameter
	unescapedQueryParam, err := url.QueryUnescape(req.URL.RawQuery)
	if err != nil {
		log.Error(req.Context(), "failed to unescape query parameters", err, log.Data{"query_param": req.URL.RawQuery})
		auditEvent.QueryParam = req.URL.RawQuery
	} else {
		auditEvent.QueryParam = unescapedQueryParam
	}

	return auditEvent
}

// retrieveIdentity requests the user and caller identity from Zebedee, using the provided client.
func retrieveIdentity(w http.ResponseWriter, req *http.Request, idClient *clientsidentity.Client, auth authorisation.Config) (ctx context.Context, status int, err error) {
	ctx = req.Context()

	florenceToken, err := getFlorenceToken(ctx, req)
	if err != nil {
		handleError(ctx, w, req, http.StatusInternalServerError, "error getting florence access token from request", err, nil)
		return ctx, http.StatusInternalServerError, err
	}

	if strings.Contains(florenceToken, ".") {
		token := florenceToken
		bearerPrefix := "Bearer "
		if strings.HasPrefix(florenceToken, bearerPrefix) {
			token = strings.TrimPrefix(florenceToken, bearerPrefix)
		}

		authorisationMiddleware, authErr := authorisation.NewFeatureFlaggedMiddleware(ctx, &auth, nil)
		if authErr != nil {
			handleError(ctx, w, req, http.StatusInternalServerError, "error getting jwtRSAPublicKeys from request", authErr, nil)
			return ctx, http.StatusInternalServerError, authErr
		}

		entityData, parseErr := authorisationMiddleware.Parse(token)
		if parseErr != nil {
			handleError(ctx, w, req, http.StatusInternalServerError, "error getting parsing token from request", parseErr, nil)
			return ctx, http.StatusInternalServerError, parseErr
		}

		if entityData != nil {
			ctx = context.WithValue(ctx, dprequest.UserIdentityKey, entityData.UserID)
			return ctx, http.StatusOK, nil
		}
	}

	log.Info(ctx, "********FLORENCE********", log.Data{
		"*********FLORENCE*********": florenceToken,
	})

	serviceAuthToken, err := getServiceAuthToken(ctx, req)
	if err != nil {
		handleError(ctx, w, req, http.StatusInternalServerError, "error getting service access token from request", err, nil)
		return ctx, http.StatusInternalServerError, err
	}

	log.Info(ctx, "********SERVICEAUTH********", log.Data{
		"*********SERVICEAUTH*********": serviceAuthToken,
	})

	// THE FLORENCE TOKEN WE USE IN THE REQUEST IS NOT COMPATIBLE WITH USER REQUESTS IN THE CHECKREQUEST FUNCTION (AND THE FUNCTIONS IT CALLS - ) BELOW

	// CheckRequest performs the call to Zebedee GET /identity and stores the values in context
	ctx, statusCode, authFailure, err := idClient.CheckRequest(req, florenceToken, serviceAuthToken)
	logData := log.Data{"auth_status_code": statusCode}
	if err != nil {
		handleError(ctx, w, req, statusCode, "identity client check request returned an error", err, logData)
		return ctx, statusCode, err
	}

	if authFailure != nil {
		handleError(ctx, w, req, statusCode, "identity client check request returned an auth error", authFailure, logData)
		return ctx, statusCode, authFailure
	}

	return ctx, http.StatusOK, nil
}

// handleError adhering to the DRY principle - clean up for failed identity requests, log the error, drain the request body and write the status code.
func handleError(ctx context.Context, w http.ResponseWriter, r *http.Request, status int, eventDetails string, err error, data log.Data) {
	log.Error(ctx, eventDetails, err, data)
	dphttp.DrainBody(r)
	w.WriteHeader(status)
}

func getFlorenceToken(ctx context.Context, req *http.Request) (string, error) {
	var florenceToken string
	// bearerPrefix := "Bearer "

	token, err := headers.GetUserAuthToken(req)
	if err == nil {
		florenceToken = token
	} else if headers.IsErrNotFound(err) {
		log.Info(ctx, "florence access token header not found attempting to find access token cookie")
		florenceToken, err = getFlorenceTokenFromCookie(ctx, req)
	}
	// if strings.HasPrefix(florenceToken, bearerPrefix) {
	// 	florenceToken = strings.TrimPrefix(florenceToken, bearerPrefix)
	// }

	return florenceToken, err
}

func getFlorenceTokenFromCookie(ctx context.Context, req *http.Request) (string, error) {
	var florenceToken string
	var err error

	c, err := req.Cookie(dprequest.FlorenceCookieKey)
	if err == nil {
		florenceToken = c.Value
	} else if err == http.ErrNoCookie {
		err = nil // we don't consider this scenario an error so we set err to nil and return an empty token
		log.Info(ctx, "florence access token cookie not found in request")
	}

	return florenceToken, err
}

func getServiceAuthToken(ctx context.Context, req *http.Request) (string, error) {
	var authToken string
	var token string
	var err error

	token, err = headers.GetServiceAuthToken(req)
	if err == nil {
		authToken = token
	} else if headers.IsErrNotFound(err) {
		err = nil // we don't consider this scenario an error so we set err to nil and return an empty token
		log.Info(ctx, "service auth token request header is not found")
	}

	return authToken, err
}
