env = {
  "FTL_SOURCE": "${HERMIT_ENV}",
  "PATH": "${HERMIT_ENV}/scripts:${HERMIT_ENV}/console/node_modules/.bin:${PATH}",
}
sources = ["env:///bin/packages", "https://github.com/cashapp/hermit-packages.git"]
