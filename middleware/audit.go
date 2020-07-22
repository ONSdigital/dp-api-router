package middleware

import (
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-router/event"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

// AuditHandler is a middleware handler that keeps track of calls that are proxied for auditing purposes.
func AuditHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auditBeforeProxying(r)
		rec := &responseRecorder{w, http.StatusOK}
		h.ServeHTTP(rec, r)
		auditAfterProxying(r, rec)
	})
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

// auditBeforeProxying is called before the HTTP request is proxied to its final destination
// it will generate an Audit event and send it as a kafka message
func auditBeforeProxying(req *http.Request) {
	auditEvent := generateAuditEvent(req)
	log.Event(req.Context(), "audit before proxying", log.INFO, log.Data{"event": auditEvent})
}

// auditAfterProxying is called after the HTTP response is received from its final destination.
// it will generate an Audit event and send it as a kafka message
func auditAfterProxying(req *http.Request, resp *responseRecorder) {
	auditEvent := generateAuditEvent(req)
	auditEvent.StatusCode = int32(resp.status)
	log.Event(req.Context(), "audit after proxying", log.INFO, log.Data{"event": auditEvent})
}

// generateAuditEvent creates an audit event with the values from request and request context, if present.
func generateAuditEvent(req *http.Request) *event.Audit {
	ctx := req.Context()
	auditEvent := &event.Audit{
		CreatedAt:  event.CreatedAtMillis(time.Now()),
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
