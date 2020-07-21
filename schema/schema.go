package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

// audit represents the schema for an audit message.
// Note that logicalType is defined by the latest avro spec,
// but is not implemented by the library used by go-ns/avro,
// so it will need to be manually translated by users of this schema
var audit = `{
  "type": "record",
  "name": "audit",
  "fields": [
    {"name": "created_at", "type": "long", "logicalType": "timestamp-millis"},
    {"name": "request_id", "type": "string", "default": ""},
    {"name": "identity", "type": "string", "default": ""},
    {"name": "collection_id", "type": "string", "default": ""},
    {"name": "path", "type": "string", "default": ""},
    {"name": "method", "type": "string", "default": ""},
    {"name": "status_code", "type": "int", "default": 0},
    {"name": "query_param", "type": "string", "default": ""}
  ]
}`

// AuditEvent is the Avro schema for Audit messages.
var AuditEvent = &avro.Schema{
	Definition: audit,
}
