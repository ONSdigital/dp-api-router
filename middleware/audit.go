package middleware

import (
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-router/event"
	"github.com/ONSdigital/dp-api-router/schema"
	kafka "github.com/ONSdigital/dp-kafka"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

// AuditHandler is a middleware handler that keeps track of calls that are proxied for auditing purposes.
func AuditHandler(auditKafkaProducer kafka.IProducer) func(h http.Handler) http.Handler {
	auditProducer := event.NewAvroProducer(auditKafkaProducer.Channels().Output, schema.AuditEvent)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Audit before proxying
			auditEvent := generateAuditEvent(r)
			if err := auditProducer.Audit(auditEvent); err != nil {
				handleAuditError(w, r, auditEvent)
				return
			}

			// Proxy the call, recording the status code
			rec := &responseRecorder{w, http.StatusOK}
			h.ServeHTTP(rec, r)

			// Audit after proxying
			auditEvent.StatusCode = int32(rec.status)
			if err := auditProducer.Audit(auditEvent); err != nil {
				handleAuditError(w, r, auditEvent)
			}
		})
	}
}

// responseRecorder implements ResponseWriter and keeps track of the status code
type responseRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader intercepts the status code and stores it in rec.status
func (rec *responseRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

//Now is a time.Now wrapper
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
	w.Write([]byte{})
	w.WriteHeader(http.StatusInternalServerError)
}
