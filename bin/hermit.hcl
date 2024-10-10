env = {
  "DBMATE_MIGRATIONS_DIR": "${HERMIT_ENV}/backend/controller/sql/schema",
  "DBMATE_NO_DUMP_SCHEMA": "true",
  "FTL_ENDPOINT": "http://127.0.0.1:8892",
  "FTL_INIT_GO_REPLACE": "github.com/TBD54566975/ftl=${HERMIT_ENV}",
  "FTL_SOURCE": "${HERMIT_ENV}",
  "OTEL_GRPC_PORT": "4317",
  "OTEL_HTTP_PORT": "4318",
  "OTEL_METRIC_EXPORT_INTERVAL": "5000",
  "PATH": "${HERMIT_ENV}/scripts:${HERMIT_ENV}/frontend/console/node_modules/.bin:${HERMIT_ENV}/frontend/vscode/node_modules/.bin:${PATH}",
}
sources = ["env:///bin/packages", "https://github.com/cashapp/hermit-packages.git"]
