# dp-api-router

A service which routes API requests to the correct services. In the future this may add additional header information
which will be used by the services.

## JSON-LD

This service is responsible for serving a JSON-LD `@context` field on configured API routes. In order to update the
JSON-LD files, follow [this guide](JSONLD.md). Once the files have been updated, consider if the secrets for this
service need to be updated to point to a new file location or not.

## Configuration

| Environment variable                     | Default                  | Description                                                                                    |
|------------------------------------------|--------------------------|------------------------------------------------------------------------------------------------|
| ALLOWED_ORIGINS                          | "http://localhost:8081"  |                                                                                                |
| BIND_ADDR                                | ":23200"                 | The host and port to bind to                                                                   |
| ENV_HOST                                 | "http://localhost:23200" | The public host for the environment the service is running on                                  |
| VERSION                                  | "v1"                     | The version of the API                                                                         |
| ENABLE_AUDIT                             | false                    |                                                                                                |
| ENABLE_OBSERVATION_API                   | false                    |                                                                                                |
| ENABLE_PRIVATE_ENDPOINTS                 | true                     | If private endpoints should be routed                                                          |
| ENABLE_V1_BETA_RESTRICTION               | false                    |                                                                                                |
| ENABLE_SESSIONS_API                      | false                    |                                                                                                |
| ENABLE_RELEASE_CALENDAR_API              | false                    | Flag to enable routing to the release calendar API                                             |
| ENABLE_CANTABULAR_METADATA_EXTRACTOR_API | false                    | Flag to enable routing to the cantabular metadata extractor API                                |
| ENABLE_ZEBEDEE_AUDIT                     | false                    |                                                                                                |
| ENABLE_NLP_SEARCH_APIS                   | false                    | Flag to enable routing to the NLP search APIs                                                  |
| CONTEXT_URL                              | ""                       | A URL to the JSON-LD context file describing the APIs                                          |
| IMPORT_API_URL                           | "http://localhost:21800" | A URL to the import api                                                                        |
| DATASET_API_URL                          | "http://localhost:22000" | A URL to the dataset api                                                                       |
| FILTER_API_URL                           | "http://localhost:22100" | A URL to the filter api                                                                        |
| RECIPE_API_URL                           | "http://localhost:22300" | A URL to the recipe api                                                                        |
| CODE_LIST_API_URL                        | "http://localhost:22400" | A URL to the code list api                                                                     |
| HIERARCHY_API_URL                        | "http://localhost:22600" | A URL to the hierarchy api                                                                     |
| SEARCH_API_URL                           | "http://localhost:23900" | A URL to the search api                                                                        |
| DIMENSION_SEARCH_API_URL                 | "http://localhost:23100" | A URL to the dimension search api                                                              |
| SESSIONS_API_URL                         | "http://localhost:24400" | A URL to the sessions api                                                                      |
| OBSERVATION_API_URL                      | "http://localhost:24500" | A URL to the observation api                                                                   |
| IMAGE_API_URL                            | "http://localhost:24700" | A URL to the image api                                                                         |
| UPLOAD_SERVICE_API_URL:                  | "http://localhost:25100" | A URL to the upload service api                                                                |
| FILES_API_URL:                           | "http://localhost:26900" | A URL to the files API                                                                         |
| DOWNLOAD_SERVICE_URL:                    | "http://localhost:23600" | A URL to the download service API                                                              |
| IDENTITY_API_URL:                        | "http://localhost:25600" | A URL to the identity api                                                                      |
| IDENTITY_API_VERSIONS                    | "v1"                     | A comma delimted string with a list of versions supported by identity api                      |
| PERMISSIONS_API_URL:                     | "http://localhost:25400" | A URL to the permissions api                                                                   |
| PERMISSIONS_API_VERSIONS                 | "v1"                     | A comma delimted string with a list of versions supported by permissions api                   |
| RELEASE_CALENDAR_API_URL                 | "http://localhost:27800" | A URL to the release calendar api                                                              |
| ZEBEDEE_URL                              | "http://localhost:8082"  | A URL to the zebedee service api                                                               |
| CANTABULAR_METADATA_EXTRACTOR_API_URL    | "http://localhost:28300" | A URL to the Cantabular metadata extractor API                                                 |
| SEARCH_SCRUBBER_API_URL                  | "http://localhost:28700" | A URL to the Cantabular metadata extractor API                                                 |
| SEARCH_SCRUBBER_API_VERSIONS             | "v1"                     | A comma delimted string with a list of versions supported by search scrubber api               |
| CATEGORY_API_URL                         | "http://localhost:28800" | A URL to the NLP Category API                                                                  |
| CATEGORY_API_VERSIONS                    | "v1"                     | A comma delimted string with a list of versions supported by search scrubber api               |
| BERLIN_API_URL                           | "http://localhost:28900" | A URL to the NLP Berlin API                                                                    |
| BERLIN_API_VERSIONS                      | "v1"                     | A comma delimted string with a list of versions supported by search scrubber api               |
| KAFKA_ADDR                               | localhost:9092           | The list of kafka hosts                                                                        |
| KAFKA_MAX_BYTES                          | 2000000                  | The maximum bytes that can be sent in an event to kafka topic                                  |
| KAFKA_VERSION                            | "1.0.2"                  | The kafka version that this service expects to connect to                                      |
| KAFKA_SEC_PROTO                          | _unset_ (only `TLS`)     | if set to `TLS`, kafka connections will use TLS                                                |
| KAFKA_SEC_CLIENT_KEY                     | _unset_                  | PEM [2] for the client key (optional, used for client auth) [1]                                |
| KAFKA_SEC_CLIENT_CERT                    | _unset_                  | PEM [2] for the client certificate (optional, used for client auth) [1]                        |
| KAFKA_SEC_CA_CERTS                       | _unset_                  | PEM [2] of CA cert chain if using private CA for the server cert [1]                           |
| KAFKA_SEC_SKIP_VERIFY                    | false                    | ignore server certificate issues if set to `true` [1]                                          |
| KAFKA_MIN_HEALTHY_BROKERS                | 0                        | The minimum number of healthy brokers, else app stays unhealthy [3]                            |
| AUDIT_TOPIC                              | audit                    | The kafka topic name for audit events                                                          |
| HEALTHCHECK_INTERVAL                     | 30s                      | The period of time between health checks                                                       |
| HEALTHCHECK_CRITICAL_TIMEOUT             | 90s                      | The period of time after which failing checks will result in critical global check             |
| SHUTDOWN_TIMEOUT                         | 5s                       | The graceful shutdown timeout (`time.Duration` format)                                         |
| HTTP_WRITE_TIMEOUT                       | dp-net default           | How long the dispatcher waits for us to write to it (`time.Duration` format)                   |
| OTEL_EXPORTER_OTLP_ENDPOINT              | localhost:4317           | Host and port for the OpenTelemetry endpoint                                                   |
| OTEL_SERVICE_NAME                        | dp-api-router            | Service name to report to telemetry tools                                                      |
| DEPRECATION_CONFIG_FILE_PATH             | _unset_                  | Optional path to a separate deprecations config file loaded at startup (see below for details) |

### Deprecation configuration

A separate configuration file can be supplied via environment variable `DEPRECATION_CONFIG_FILE_PATH` containing
definitions of deprecated paths. If this environment variable is unset or is an empty string, then the deprecation 
middleware is disabled. Otherwise, these paths will then be handled by middleware that adds deprecation headers as well
as returning a custom error message and NOT_FOUND(404) status during defined outage periods.

The format of the configuration file is as follows…

```json
[
  {
    "paths": [
      "/path1",
      "/path2/"
    ],
    "date": "2024-09-10T10:00:00Z",
    "sunset": "2024-10-14",
    "outages": [
      "24h@2024-12-31",
      "12h@2025-01-01T10:00:00Z",
      "2025-02-03 16:00:00"
    ],
    "link": "https://developer.ons.gov.uk/retirement/v0api/",
    "msg": "This API is being decommissioned …"
  }
]
```

Where the fields are defined as…

- `paths` : a list of paths conforming
  to [the http.ServeMux patterns](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)
- `date` : the date the paths became deprecated
- `sunset` : the start date of the sunset period
- `outages` : a list of durations (optional) and dates/times for periods where matching requests should be rejected as
  NOT_FOUND(44). Note, valid time units for the duration are "ns", "us" (or "µs"), "ms", "s", "m", "h" (as
  per [Go's `time.ParseDuration` function](https://pkg.go.dev/time#ParseDuration)). If the duration is absent, the outage will apply forever from the
  specified time onwards.
- `link` : the URL to be returned in the `sunset` relational link header
- `msg` : the message to be returned in responses to matching requests during the outage periods.

All dates are parsed according to either RFC3339 ("2006-01-02T15:04:05Z07:00"), DateOnly ("2006-01-02") or DateTime
("2006-01-02 15:04:05") as defined in [Go's `time` package](https://pkg.go.dev/time#pkg-constants)

Notes:

1. Ignored unless using TLS (i.e. `KAFKA_SEC_PROTO` has a value enabling TLS)

2. PEM values are identified as those starting with `-----BEGIN`
   and can use `\n` (sic) instead of newlines (they will be converted to newlines before use).
   Any other value will be treated as a path to the given PEM file.

3. the default is `0` which means "use library default" - recommended to change this _only in development_ (e.g. when
   running only one broker).
