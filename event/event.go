package event

import "time"

// Audit provides an avro structure for an audit event
// Note: createdAt should be a time.Time, but it is not currently supported by the avro library
type Audit struct {
	CreatedAt    int64  `avro:"created_at"`
	RequestID    string `avro:"request_id"`
	Identity     string `avro:"identity"`
	CollectionID string `avro:"collection_id"`
	Path         string `avro:"path"`
	Method       string `avro:"method"`
	StatusCode   int32  `avro:"status_code"`
	QueryParam   string `avro:"query_param"`
}

// CreatedAtTime returns a time.Time representation of the CreatedAt field of an Audit struct
func (a *Audit) CreatedAtTime() time.Time {
	var sec, nanosec int64
	sec = a.CreatedAt / 1e3
	nanosec = (a.CreatedAt % 1e3) * 1e6
	return time.Unix(sec, nanosec).UTC()
}

// CreatedAtMillis returns the number of milliseconds since Unix time reference
func CreatedAtMillis(t time.Time) int64 {
	return t.Unix()*1e3 + int64(t.Nanosecond()/1e6)
}
