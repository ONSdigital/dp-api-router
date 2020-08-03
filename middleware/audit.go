package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-api-router/event"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/request"
	"github.com/ONSdigital/log.go/log"
)

// AuditHandler is a middleware handler that keeps track of calls for auditing purposes,
// before and after proxying calling the downstream service.
// It obtains the user and caller information by calling Zebedee GET /identity
func AuditHandler(auditProducer *event.AvroProducer, cli dphttp.Clienter, zebedeeURL string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Inbound audit event (before proxying).
			auditEvent := generateAuditEvent(r)

			// Retrieve Identity from Zebedee, which is stored in context.
			// if it fails, try to audit with the statusCode before returning
			ctx, statusCode, err := retrieveIdentity(w, r, cli, zebedeeURL)
			if err != nil {
				// error already handled in retrieveIdentity. Try to audit it.
				auditEvent.StatusCode = int32(statusCode)
				if err := auditProducer.Audit(auditEvent); err != nil {
					log.Event(ctx, "inbound audit event could not be sent", log.ERROR, log.Data{"event": auditEvent})
				}
				return
			}
			r = r.WithContext(ctx)

			// Add identity to audit event. User identity takes priority over service identity.
			// If no identity is available, then try to audit without identity and fail the request.
			userIdentity := common.User(ctx)
			serviceIdentity := common.Caller(ctx)
			if userIdentity != "" {
				auditEvent.Identity = userIdentity
			} else if serviceIdentity != "" {
				auditEvent.Identity = serviceIdentity
			} else {
				handleError(ctx, w, r, http.StatusUnauthorized, "", err, log.Data{"event": auditEvent})
				auditEvent.StatusCode = int32(http.StatusUnauthorized)
				if err := auditProducer.Audit(auditEvent); err != nil {
					log.Event(ctx, "inbound audit event could not be sent", log.ERROR, log.Data{"event": auditEvent})
				}
				return
			}

			// Acceptable request. Audit it before proxying.
			if err := auditProducer.Audit(auditEvent); err != nil {
				handleError(ctx, w, r, http.StatusInternalServerError, "inbound audit event could not be sent", err, log.Data{"event": auditEvent})
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
				handleError(ctx, w, r, http.StatusInternalServerError, "outbound audit event could not be sent", err, log.Data{"event": auditEvent})
				return
			}

			// Copy the intercepted body and header to the original response writer
			w.WriteHeader(rec.statusCode)
			io.Copy(w, rec.body)

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

// generateAuditEvent creates an audit event with the values from request and request context, if present.
func generateAuditEvent(req *http.Request) *event.Audit {
	auditEvent := &event.Audit{
		CreatedAt:  event.CreatedAtMillis(Now()),
		Path:       req.URL.Path,
		Method:     req.Method,
		QueryParam: req.URL.RawQuery,
	}
	auditEvent.RequestID = common.GetRequestId(req.Context())
	if colID := req.Header.Get(dphttp.CollectionIDHeaderKey); colID != "" {
		auditEvent.CollectionID = colID
	}
	return auditEvent
}

// retrieveIdentity requests the user and caller identity from Zebedee, using hte provided client and url.
func retrieveIdentity(w http.ResponseWriter, req *http.Request, cli dphttp.Clienter, zebedeeURL string) (ctx context.Context, status int, err error) {
	ctx = req.Context()
	log.Event(ctx, "executing identity check for auditing purposes")

	idClient := clientsidentity.NewAPIClient(cli, zebedeeURL)

	florenceToken, err := getFlorenceToken(ctx, req)
	if err != nil {
		handleError(ctx, w, req, http.StatusInternalServerError, "error getting florence access token from request", err, nil)
		return ctx, http.StatusInternalServerError, err
	}

	serviceAuthToken, err := getServiceAuthToken(ctx, req)
	if err != nil {
		handleError(ctx, w, req, http.StatusInternalServerError, "error getting service access token from request", err, nil)
		return ctx, http.StatusInternalServerError, err
	}

	// CheckRequest performs the call to Zebedee GET /identity and stores the values  in context
	ctx, statusCode, authFailure, err := idClient.CheckRequest(req, florenceToken, serviceAuthToken)
	logData := log.Data{"auth_status_code": statusCode}
	if err != nil {
		handleError(ctx, w, req, statusCode, "identity client check request returned an error", err, logData)
		return ctx, statusCode, err
	}

	if authFailure != nil {
		handleError(ctx, w, req, statusCode, "identity client check request returned an auth error", authFailure, logData)
		log.Event(ctx, "identity client check request returned an auth error", log.Error(authFailure), logData)
		return ctx, statusCode, authFailure
	}

	return ctx, http.StatusOK, nil
}

// handleError adhering to the DRY principle - clean up for failed identity requests, log the error, drain the request body and write the status code.
func handleError(ctx context.Context, w http.ResponseWriter, r *http.Request, status int, event string, err error, data log.Data) {
	log.Event(ctx, event, log.Error(err), log.ERROR, data)
	request.DrainBody(r)
	w.WriteHeader(status)
}

func getFlorenceToken(ctx context.Context, req *http.Request) (string, error) {
	var florenceToken string

	token, err := headers.GetUserAuthToken(req)
	if err == nil {
		florenceToken = token
	} else if headers.IsErrNotFound(err) {
		log.Event(ctx, "florence access token header not found attempting to find access token cookie")
		florenceToken, err = getFlorenceTokenFromCookie(ctx, req)
	}

	return florenceToken, err
}

func getFlorenceTokenFromCookie(ctx context.Context, req *http.Request) (string, error) {
	var florenceToken string
	var err error

	c, err := req.Cookie(dphttp.FlorenceCookieKey)
	if err == nil {
		florenceToken = c.Value
	} else if err == http.ErrNoCookie {
		err = nil // we don't consider this scenario an error so we set err to nil and return an empty token
		log.Event(ctx, "florence access token cookie not found in request")
	}

	return florenceToken, err
}

func getServiceAuthToken(ctx context.Context, req *http.Request) (string, error) {
	var authToken string

	token, err := headers.GetServiceAuthToken(req)
	if err == nil {
		authToken = token
	} else if headers.IsErrNotFound(err) {
		err = nil // we don't consider this scenario an error so we set err to nil and return an empty token
		log.Event(ctx, "service auth token request header is not found")
	}

	return authToken, err
}

// Now is a time.Now wrapper
var Now = func() time.Time {
	return time.Now()
}
