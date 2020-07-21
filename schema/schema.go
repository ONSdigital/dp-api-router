package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

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
    {"name": "status_code", "type": "int"},
    {"name": "query_param", "type": "string", "default": ""}
  ]
}`

// AuditEvent is the Avro schema for Audit messages.
var AuditEvent = &avro.Schema{
	Definition: audit,
}
