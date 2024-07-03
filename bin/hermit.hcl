env = {
  "FTL_ENDPOINT": "http://localhost:8892",
  "FTL_SOURCE": "${HERMIT_ENV}",
  "OTEL_METRIC_EXPORT_INTERVAL": "5000",
  "PATH": "${HERMIT_ENV}/scripts:${HERMIT_ENV}/frontend/node_modules/.bin:${HERMIT_ENV}/extensions/vscode/node_modules/.bin:${PATH}",
  "FTL_INIT_GO_REPLACE": "github.com/TBD54566975/ftl=${HERMIT_ENV}",

}
sources = ["env:///bin/packages", "https://github.com/cashapp/hermit-packages.git"]
