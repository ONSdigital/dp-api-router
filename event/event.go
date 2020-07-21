package event

import "time"

// Audit provides an avro structure for an audit event
type Audit struct {
	CreatedAt    time.Time `avro:"created_at"`
	RequestID    string    `avro:"request_id"`
	Identity     string    `avro:"identity"`
	CollectionID string    `avro:"collection_id"`
	Path         string    `avro:"path"`
	Method       string    `avro:"method"`
	StatusCode   int       `avro:"status_code"`
	QueryParam   string    `avro:"query_param"`
}
