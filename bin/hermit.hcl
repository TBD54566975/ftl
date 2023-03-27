env = {
  "FTL_SOURCE": "${HERMIT_ENV}",
  "PATH": "${HERMIT_ENV}/scripts:${PATH}",
}
sources = ["env:///bin/packages", "https://github.com/cashapp/hermit-packages.git"]
