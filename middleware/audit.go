package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-router/event"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

// ignoredPaths will not be audited
var ignoredPaths = map[string]bool{
	"/health":      true,
	"/healthcheck": true,
}

// AuditHandler is a middleware handler that keeps track of calls that are proxied for auditing purposes.
func AuditHandler(auditProducer *event.AvroProducer) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// check if path needs to be ignored (not audited)
			if _, found := ignoredPaths[r.URL.Path]; found {
				h.ServeHTTP(w, r)
				return
			}

			// Audit before proxying
			auditEvent := generateAuditEvent(r)
			if err := auditProducer.Audit(auditEvent); err != nil {
				handleAuditError(w, r, auditEvent)
				return
			}

			// Proxy the call with our responseRecorder
			rec := &responseRecorder{w, http.StatusOK, &bytes.Buffer{}}
			h.ServeHTTP(rec, r)

			// Audit after proxying. We need to marshal first, and then send if there is no other error in the process
			auditEvent.StatusCode = int32(rec.statusCode)
			eventBytes, err := auditProducer.Marshal(auditEvent)
			if err != nil {
				handleAuditError(w, r, auditEvent)
				return
			}

			// Copy the intercepted body and header to the original response writer
			w.WriteHeader(rec.statusCode)
			if _, err := io.Copy(w, rec.body); err != nil {
				handleResponseCopyError(w, r, auditEvent)
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

// Now is a time.Now wrapper
var Now = func() time.Time {
	return time.Now()
}

// generateAuditEvent creates an audit event with the values from request and request context, if present.
func generateAuditEvent(req *http.Request) *event.Audit {
	ctx := req.Context()
	auditEvent := &event.Audit{
		CreatedAt:  event.CreatedAtMillis(Now()),
		Path:       req.URL.Path,
		Method:     req.Method,
		QueryParam: req.URL.RawQuery,
	}
	if ctx.Value(dphttp.RequestIdKey) != nil {
		auditEvent.RequestID = ctx.Value(dphttp.RequestIdKey).(string)
	}
	if ctx.Value(dphttp.CallerIdentityKey) != nil {
		auditEvent.Identity = ctx.Value(dphttp.CallerIdentityKey).(string)
	}
	if ctx.Value(dphttp.CollectionIDHeaderKey) != nil {
		auditEvent.CollectionID = ctx.Value(dphttp.CollectionIDHeaderKey).(string)
	}
	return auditEvent
}

// handleAuditError logs the audit event that failed to be sent and sets the HTTP error response and empty body
func handleAuditError(w http.ResponseWriter, r *http.Request, auditEvent *event.Audit) {
	log.Event(r.Context(), "audit event could not be sent", log.ERROR, log.Data{"event": auditEvent})
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte{})
}

// handleResponseCopyError logs the error and sets the HTTP error response and empty body
func handleResponseCopyError(w http.ResponseWriter, r *http.Request, auditEvent *event.Audit) {
	log.Event(r.Context(), "error copying the intercepted responseWriter into the final responseWriter, after successfully having generated the audit event",
		log.ERROR, log.Data{"event": auditEvent})
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte{})
}
