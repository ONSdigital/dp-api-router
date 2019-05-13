dp-api-router
=====================
A service which routes API requests to the correct services. In the future this may add additional header information which will be used by the services.

### JSON-LD

This service is responsible for serving a JSON-LD `@context` field on configured API routes. In order to update the JSON-LD files, follow [this guide](JSONLD.md). Once the files have been updated, consider if the secrets for this service need to be updated to point to a new file location or not.

### Configuration

| Environment variable       | Default                                   | Description
| -------------------------- | ----------------------------------------- | -----------
| BIND_ADDR                  | ":23200"                                  | The host and port to bind to
| VERSION                    | "v1"                                      | The version of the API
| ENABLE_PRIVATE_ENDPOINTS   | true                                      | If private endpoints should be routed
| HIERARCHY_API_URL          | "http://localhost:22600"                  | A URL to the hierarchy api
| FILTER_API_URL             | "http://localhost:22100"                  | A URL to the filter api
| DATASET_API_URL            | "http://localhost:22000"                  | A URL to the dataset api
| CODE_LIST_API_URL          | "http://localhost:22400"                  | A URL to the code list api
| RECIPE_API_URL             | "http://localhost:22300"                  | A URL to the recipe api
| IMPORT_API_URL             | "http://localhost:21800"                  | A URL to the import api
| SEARCH_API_URL             | "http://localhost:23100"                  | A URL to the search api
| API_POC_URL                | "http://localhost:3000"                   | A URL to the poc api
| CONTEXT_URL                | ""                   			 | A URL to the JSON-LD context file describing the APIs
| SHUTDOWN_TIMEOUT           | 5s                                        | The graceful shutdown timeout (`time.Duration` format)
| ENV_HOST                   | "http://localhost:23200"                  | The public host for the environment the service is running on
