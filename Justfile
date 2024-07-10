set positional-arguments
set shell := ["bash", "-c"]

WATCHEXEC_ARGS := "-d 1s -e proto -e go -e sql -f sqlc.yaml"
RELEASE := "build/release"
VERSION := `git describe --tags --always | sed -e 's/^v//'`
KT_RUNTIME_OUT := "kotlin-runtime/ftl-runtime/target/ftl-runtime-1.0-SNAPSHOT.jar"
KT_RUNTIME_RUNNER_TEMPLATE_OUT := "build/template/ftl/jars/ftl-runtime.jar"
RUNNER_TEMPLATE_ZIP := "backend/controller/scaling/localscaling/template.zip"
TIMESTAMP := `date +%s`
SCHEMA_OUT := "backend/protos/xyz/block/ftl/v1/schema/schema.proto"
ZIP_DIRS := "go-runtime/compile/build-template go-runtime/compile/external-module-template go-runtime/compile/main-work-template common-runtime/scaffolding go-runtime/scaffolding kotlin-runtime/scaffolding kotlin-runtime/external-module-template"
FRONTEND_OUT := "frontend/dist/index.html"
EXTENSION_OUT := "extensions/vscode/dist/extension.js"
PROTOS_IN := "backend/protos/xyz/block/ftl/v1/schema/schema.proto backend/protos/xyz/block/ftl/v1/console/console.proto backend/protos/xyz/block/ftl/v1/ftl.proto backend/protos/xyz/block/ftl/v1/schema/runtime.proto"
PROTOS_OUT := "backend/protos/xyz/block/ftl/v1/console/console.pb.go backend/protos/xyz/block/ftl/v1/ftl.pb.go backend/protos/xyz/block/ftl/v1/schema/runtime.pb.go backend/protos/xyz/block/ftl/v1/schema/schema.pb.go frontend/src/protos/xyz/block/ftl/v1/console/console_pb.ts frontend/src/protos/xyz/block/ftl/v1/ftl_pb.ts frontend/src/protos/xyz/block/ftl/v1/schema/runtime_pb.ts frontend/src/protos/xyz/block/ftl/v1/schema/schema_pb.ts"

_help:
  @just -l

# Run errtrace on Go files to add stacks
errtrace:
  git ls-files -z -- '*.go' | grep -zv /_ | xargs -0 errtrace -w && go mod tidy

# Clean the build directory
clean:
  rm -rf build
  rm -rf frontend/node_modules
  find . -name '*.zip' -exec rm {} \;
  mvn -f kotlin-runtime/ftl-runtime clean

# Live rebuild the ftl binary whenever source changes.
live-rebuild:
  watchexec {{WATCHEXEC_ARGS}} -- "just build-sqlc && just build ftl"

# Run "ftl dev" with live-reloading whenever source changes.
dev *args:
  watchexec -r {{WATCHEXEC_ARGS}} -- "just build-sqlc && ftl dev {{args}}"

# Build everything
build-all: build-protos-unconditionally build-frontend build-generate build-sqlc build-zips lsp-generate
  @just build ftl ftl-controller ftl-runner ftl-initdb

# Run "go generate" on all packages
build-generate:
  @mk backend/schema/aliaskind_enumer.go : backend/schema/metadataalias.go -- go generate -x ./backend/schema
  @mk internal/log/log_level_string.go : internal/log/api.go -- go generate -x ./internal/log

# Build command-line tools
build +tools: build-protos build-zips build-frontend
  #!/bin/bash
  shopt -s extglob

  if [ "${FTL_DEBUG:-}" = "true" ]; then
    for tool in $@; do go build -o "{{RELEASE}}/$tool" -tags release -gcflags=all="-N -l" -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./cmd/$tool"; done
  else
    for tool in $@; do mk "{{RELEASE}}/$tool" : !(build|integration) -- go build -o "{{RELEASE}}/$tool" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./cmd/$tool"; done
  fi

# Build all backend binaries
build-backend:
  just build ftl ftl-controller ftl-runner

export DATABASE_URL := "postgres://postgres:secret@localhost:15432/ftl?sslmode=disable"

# Explicitly initialise the database
init-db:
  dbmate drop || true
  dbmate create
  dbmate --migrations-dir backend/controller/sql/schema up

# Regenerate SQLC code (requires init-db to be run first)
build-sqlc:
  @mk backend/controller/sql/{db.go,models.go,querier.go,queries.sql.go} backend/controller/{cronjobs}/sql/{db.go,models.go,querier.go,queries.sql.go} common/configuration/sql/{db.go,models.go,querier.go,queries.sql.go} : backend/controller/sql/queries.sql backend/controller/{cronjobs}/sql/queries.sql common/configuration/sql/queries.sql backend/controller/sql/schema sqlc.yaml -- "just init-db && sqlc generate"

# Build the ZIP files that are embedded in the FTL release binaries
build-zips: build-kt-runtime
  @for dir in {{ZIP_DIRS}}; do (cd $dir && mk ../$(basename ${dir}).zip : . -- "rm -f $(basename ${dir}.zip) && zip -q --symlinks -r ../$(basename ${dir}).zip ."); done

# Rebuild frontend
build-frontend: npm-install
  @mk {{FRONTEND_OUT}} : frontend/src -- "cd frontend && npm run build"

# Rebuild VSCode extension
build-extension: npm-install
  @mk {{EXTENSION_OUT}} : extensions/vscode/src extensions/vscode/package.json -- "cd extensions/vscode && rm -f ftl-*.vsix && npm run compile"

# Install development version of VSCode extension
install-extension: build-extension
  @cd extensions/vscode && vsce package && code --install-extension ftl-*.vsix

# Build and package the VSCode extension
package-extension: build-extension
  @cd extensions/vscode && vsce package

# Publish the VSCode extension
publish-extension: package-extension
  @cd extensions/vscode && vsce publish

# Kotlin runtime is temporarily disabled; these instructions create a dummy zip in place of the kotlin runtime jar for
# the runner.
build-kt-runtime:
  @mkdir -p build/template/ftl && touch build/template/ftl/temp.txt
  @cd build/template && zip -q --symlinks -r ../../{{RUNNER_TEMPLATE_ZIP}} .

# Install Node dependencies
npm-install:
  @mk frontend/node_modules : frontend/package.json -- "cd frontend && npm install"
  @mk extensions/vscode/node_modules : extensions/vscode/package.json extensions/vscode/src -- "cd extensions/vscode && npm install"

# Regenerate protos
build-protos: npm-install
  @mk {{SCHEMA_OUT}} : backend/schema -- "ftl-schema > {{SCHEMA_OUT}} && buf format -w && buf lint"
  @mk {{PROTOS_OUT}} : {{PROTOS_IN}} -- "cd backend/protos && buf generate"

# Unconditionally rebuild protos
build-protos-unconditionally: npm-install
  ftl-schema > {{SCHEMA_OUT}} && buf format -w && buf lint
  cd backend/protos && buf generate

# Run integration test(s)
integration-tests *test:
  #!/bin/bash
  set -euo pipefail
  testName=${1:-}
  go test -fullpath -count 1 -v -tags integration -run "$testName" -p 1 $(find . -type f -name '*_test.go' -print0 | xargs -0 grep -r -l "$testName" | xargs grep -l '//go:build integration' | xargs -I {} dirname './{}')

# Run README doc tests
test-readme *args:
  mdcode run {{args}} README.md -- bash test.sh

# Run "go mod tidy" on all packages including tests
tidy:
  find . -name go.mod -execdir go mod tidy \;

# Check for changes in existing SQL migrations compared to main
ensure-frozen-migrations:
  scripts/ensure-frozen-migrations

# Run backend tests
test-backend:
  @gotestsum --hide-summary skipped --format-hide-empty-pkg -- -short -fullpath ./...

# Lint the frontend
lint-frontend: build-frontend
  @cd frontend && npm run lint && tsc

# Lint the backend
lint-backend:
  @golangci-lint run --new-from-rev=$(git merge-base origin/main HEAD) ./...

# Run live docs server
docs:
  git submodule update --init --recursive
  cd docs && zola serve

# Generate LSP hover help text
lsp-generate:
  @mk lsp/hoveritems.go : lsp docs/content -- "scripts/ftl-gen-lsp"

# Run `ftl dev` providing a Delve endpoint for attaching a debugger.
debug *args:
  #!/bin/bash
  set -euo pipefail

  cleanup() {
    if [ -n "${dlv_pid:-}" ] && kill -0 "$dlv_pid" 2>/dev/null; then
      kill "$dlv_pid"
    fi
  }
  trap cleanup EXIT

  FTL_DEBUG=true just build ftl
  dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec "{{RELEASE}}/ftl" -- dev {{args}} &
  dlv_pid=$!
  wait "$dlv_pid"

# Run otel collector in a docker container to stream local (i.e. from ftl dev) signals to
# the terminal tab where this is running. To stop the stream, run `just otel-stop` in a
# second terminal window.
otel-stream:
  docker run \
    -p 4317:4317 \
    -p 55679:55679 \
    otel/opentelemetry-collector:0.104.0

otelContainerID := `docker ps -f ancestor=otel/opentelemetry-collector:0.104.0 | tail -1 | cut -d " " -f1`
otel-stop:
  docker stop "{{otelContainerID}}"

# Run `ftl dev` with the given args after setting the necessary envar
otel-dev *args:
  export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
  ftl dev {{args}}
