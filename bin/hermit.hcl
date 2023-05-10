env = {
  "FTL_SOURCE": "${HERMIT_ENV}",
  "GOOSE_DBSTRING": "postgres://postgres:secret@127.0.0.1:5432/ftl",
  "GOOSE_DRIVER": "postgres",
  "PATH": "${HERMIT_ENV}/scripts:${HERMIT_ENV}/console/node_modules/.bin:${PATH}",
}
sources = ["env:///bin/packages", "https://github.com/cashapp/hermit-packages.git"]
