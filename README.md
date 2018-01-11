dp-api-router
=====================
A service which routes API requests to the correct services. In the future this may add additional header information, which will be used by the services.

### Configuration

| Environment variable       | Default                                   | Description
| -------------------------- | ----------------------------------------- | -----------
| BIND_ADDR                  | ":8081"                                   | The host and port to bind to
| HIERARCHY_API_URL          | "http://localhost:22600"                  | A URL to the hierarchy api
| FILTER_API_URL             | "http://localhost:22100"                  | A URL to the filter api
| DATASET_API_URL            | "http://localhost:22000"                  | A URL to the dataset api
| CODELIST_API_URL           | "http://localhost:22400"                  | A URL to the codelist api
| RECIPE_API_URL             | "http://localhost:22300"                  | A URL to the recipe api
| IMPORT_API_URL             | "http://localhost:21800"                  | A URL to the import api
| SHUTDOWN_TIMEOUT           | 5s                                        | The graceful shutdown timeout (`time.Duration` format)
