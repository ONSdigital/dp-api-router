dp-api-router
=====================
A service which routes API requests to the correct services. In the future this may add additional header information which will be used by the services.

### JSON-LD

This service is responsible for serving a JSON-LD `@context` field on configured API routes. In order to update the JSON-LD files, follow [this guide](JSONLD.md). Once the files have been updated, consider if the secrets for this service need to be updated to point to a new file location or not.

### Configuration

| Environment variable         | Default                                   | Description
| ---------------------------- | ----------------------------------------- | -----------
| ALLOWED_ORIGINS              | "http://localhost:8081"                   | 
| BIND_ADDR                    | ":23200"                                  | The host and port to bind to
| ENV_HOST                     | "http://localhost:23200"                  | The public host for the environment the service is running on
| VERSION                      | "v1"                                      | The version of the API
| ENABLE_AUDIT                 | false                                     | 
| ENABLE_OBSERVATION_API       | false                                     | 
| ENABLE_PRIVATE_ENDPOINTS     | true                                      | If private endpoints should be routed
| ENABLE_V1_BETA_RESTRICTION   | false                                     | 
| ENABLE_SESSIONS_API          | false                                     | 
| ENABLE_TOPIC_API          | false                                     | 
| ENABLE_ZEBEDEE_AUDIT         | false                                     | 
| API_POC_URL                  | "http://localhost:3000"                   | A URL to the poc api
| CODE_LIST_API_URL            | "http://localhost:22400"                  | A URL to the code list api
| CONTEXT_URL                  | ""                                        | A URL to the JSON-LD context file describing the APIs
| DATASET_API_URL              | "http://localhost:22000"                  | A URL to the dataset api
| DIMENSION_SEARCH_API_URL     | "http://localhost:23100"                  | A URL to the dimension search api
| FILTER_API_URL               | "http://localhost:22100"                  | A URL to the filter api
| HIERARCHY_API_URL            | "http://localhost:22600"                  | A URL to the hierarchy api
| IMAGE_API_URL                | "http://localhost:24700"                  | A URL to the image api
| IMPORT_API_URL               | "http://localhost:21800"                  | A URL to the import api
| OBSERVATION_API_URL          | "http://localhost:24500"                  | A URL to the observation api
| RECIPE_API_URL               | "http://localhost:22300"                  | A URL to the recipe api
| SEARCH_API_URL               | "http://localhost:23100"                  | A URL to the search api
| SESSIONS_API_URL             | "http://localhost:24400"                  | A URL to the sessions api
| UPLOAD_SERVICE_API_URL:      | "http://localhost:25100"                  | A URL to the upload service api
| ZEBEDEE_URL                  | "http://localhost:8082"                   | A URL to the zebedee service api
| KAFKA_ADDR                   | localhost:9092                            | The list of kafka hosts
| KAFKA_MAX_BYTES              | 2000000                                   | The maximum bytes that can be sent in an event to kafka topic
| KAFKA_VERSION                | "1.0.2"                                   | The kafka version that this service expects to connect to
| AUDIT_TOPIC                  | audit                                     | The kafka topic name for audit events 
| HEALTHCHECK_INTERVAL         | 30s                                       | The period of time between health checks
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                                       | The period of time after which failing checks will result in critical global check
| SHUTDOWN_TIMEOUT             | 5s                                        | The graceful shutdown timeout (`time.Duration` format)