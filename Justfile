set positional-arguments
set shell := ["bash", "-c"]

WATCHEXEC_ARGS := "-d 1s -e proto -e go -e sql -f sqlc.yaml"
RELEASE := "build/release"
VERSION := `git describe --tags --always | sed -e 's/^v//'`
RUNNER_TEMPLATE_ZIP := "backend/controller/scaling/localscaling/template.zip"
TIMESTAMP := `date +%s`
SCHEMA_OUT := "backend/protos/xyz/block/ftl/v1/schema/schema.proto"
ZIP_DIRS := "go-runtime/compile/build-template go-runtime/compile/external-module-template go-runtime/compile/main-work-template internal/projectinit/scaffolding go-runtime/scaffolding jvm-runtime/java/scaffolding jvm-runtime/kotlin/scaffolding"
CONSOLE_ROOT := "frontend/console"
FRONTEND_OUT := "{{CONSOLE_ROOT}}/dist/index.html"
EXTENSION_OUT := "frontend/vscode/dist/extension.js"
PROTOS_IN := "backend/protos/xyz/block/ftl/v1/schema/schema.proto backend/protos/xyz/block/ftl/v1/console/console.proto backend/protos/xyz/block/ftl/v1/ftl.proto backend/protos/xyz/block/ftl/v1/schema/runtime.proto"
PROTOS_OUT := "backend/protos/xyz/block/ftl/v1/console/console.pb.go backend/protos/xyz/block/ftl/v1/ftl.pb.go backend/protos/xyz/block/ftl/v1/schema/runtime.pb.go backend/protos/xyz/block/ftl/v1/schema/schema.pb.go {{CONSOLE_ROOT}}/src/protos/xyz/block/ftl/v1/console/console_pb.ts {{CONSOLE_ROOT}}/src/protos/xyz/block/ftl/v1/ftl_pb.ts {{CONSOLE_ROOT}}/src/protos/xyz/block/ftl/v1/schema/runtime_pb.ts {{CONSOLE_ROOT}}/src/protos/xyz/block/ftl/v1/schema/schema_pb.ts"

_help:
  @just -l

k8s command="_help" *args="":
  just deployment/{{command}} {{args}}

# Run errtrace on Go files to add stacks
errtrace:
  git ls-files -z -- '*.go' | grep -zv /_ | xargs -0 errtrace -w && go mod tidy

# Clean the build directory
clean:
  rm -rf build
  rm -rf node_modules
  find . -name '*.zip' -exec rm {} \;
  mvn -f jvm-runtime/ftl-runtime clean

# Live rebuild the ftl binary whenever source changes.
live-rebuild:
  watchexec {{WATCHEXEC_ARGS}} -- "just build-sqlc && just build ftl"

# Run "ftl dev" with live-reloading whenever source changes.
dev *args:
  watchexec -r {{WATCHEXEC_ARGS}} -- "just build-sqlc && ftl dev {{args}}"

# Build everything
build-all: build-protos-unconditionally build-frontend build-generate build-sqlc build-zips lsp-generate build-java
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

build-java *args:
  mvn -f jvm-runtime/ftl-runtime install {{args}}

export DATABASE_URL := "postgres://postgres:secret@localhost:15432/ftl?sslmode=disable"

# Explicitly initialise the database
init-db:
  dbmate drop || true
  dbmate create
  dbmate --migrations-dir backend/controller/sql/schema up

# Regenerate SQLC code (requires init-db to be run first)
build-sqlc:
  @mk backend/controller/sql/{db.go,models.go,querier.go,queries.sql.go} backend/controller/cronjobs/sql/{db.go,models.go,querier.go,queries.sql.go} internal/configuration/sql/{db.go,models.go,querier.go,queries.sql.go} : backend/controller/sql/queries.sql backend/controller/sql/async_queries.sql backend/controller/cronjobs/sql/queries.sql internal/configuration/sql/queries.sql backend/controller/sql/schema sqlc.yaml -- "just init-db && sqlc generate"

# Build the ZIP files that are embedded in the FTL release binaries
build-zips: build-kt-runtime
  @for dir in {{ZIP_DIRS}}; do (cd $dir && mk ../$(basename ${dir}).zip : . -- "rm -f $(basename ${dir}.zip) && zip -q --symlinks -r ../$(basename ${dir}).zip ."); done

# Rebuild frontend
build-frontend: pnpm-install
  @mk {{FRONTEND_OUT}} : {{CONSOLE_ROOT}}/src -- "cd {{CONSOLE_ROOT}} && pnpm run build"

# Rebuild VSCode extension
build-extension: pnpm-install
  @mk {{EXTENSION_OUT}} : frontend/vscode/src frontend/vscode/package.json -- "cd frontend/vscode && rm -f ftl-*.vsix && pnpm run compile"

# Install development version of VSCode extension
install-extension: build-extension
  @cd frontend/vscode && vsce package && code --install-extension ftl-*.vsix

# Build and package the VSCode extension
package-extension: build-extension
  @cd frontend/vscode && vsce package --no-dependencies

# Publish the VSCode extension
publish-extension: package-extension
  @cd frontend/vscode && vsce publish --no-dependencies

build-intellij-plugin:
  @cd frontend/intellij && gradle buildPlugin

# Kotlin runtime is temporarily disabled; these instructions create a dummy zip in place of the kotlin runtime jar for
# the runner.
build-kt-runtime:
  @mkdir -p build/template/ftl && touch build/template/ftl/temp.txt
  @cd build/template && zip -q --symlinks -r ../../{{RUNNER_TEMPLATE_ZIP}} .

# Format console code.
format-frontend:
  cd {{CONSOLE_ROOT}} && pnpm run lint:fix

# Install Node dependencies using pnpm
# Retry a few times in case of network issues
pnpm-install:
  @for i in {1..3}; do pnpm install && break || sleep 5; done

# Regenerate protos
build-protos: pnpm-install
  @mk {{SCHEMA_OUT}} : backend/schema -- "ftl-schema > {{SCHEMA_OUT}} && buf format -w && buf lint"
  @mk {{PROTOS_OUT}} : {{PROTOS_IN}} -- "cd backend/protos && buf generate"

# Unconditionally rebuild protos
build-protos-unconditionally: pnpm-install
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
  @scripts/ensure-frozen-migrations

# Run backend tests
test-backend:
  @gotestsum --hide-summary skipped --format-hide-empty-pkg -- -short -fullpath ./...

test-scripts:
  GIT_COMMITTER_NAME="CI" \
    GIT_COMMITTER_EMAIL="no-reply@tbd.email" \
    GIT_AUTHOR_NAME="CI" \
    GIT_AUTHOR_EMAIL="no-reply@tbd.email" \
    scripts/tests/test-ensure-frozen-migrations.sh

test-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && pnpm run test

e2e-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && npx playwright install --with-deps && pnpm run e2e

# Lint the frontend
lint-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && pnpm run lint && tsc

# Lint the backend
lint-backend:
  @golangci-lint run --new-from-rev=$(git merge-base origin/main HEAD) ./...
  @lint-commit-or-rollback ./backend/...

lint-scripts:
	@shellcheck -f gcc -e SC2016 $(find scripts -type f -not -path scripts/tests) | to-annotation

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
# the terminal tab where this is running. To start FTL, opepn another terminal tab and run
# `just otel-dev` with any args you would pass to `ftl dev`. To stop the otel stream, run
# `just otel-stop` in a third terminal tab.
otel-stream:
  docker run \
    -p ${OTEL_GRPC_PORT}:${OTEL_GRPC_PORT} \
    -p 55679:55679 \
    otel/opentelemetry-collector:0.104.0 2>&1 | sed 's/\([A-Z].* #\)/\
  \1/g'

# Stop the docker container running otel.
otel-container-id := `docker ps -f ancestor=otel/opentelemetry-collector:0.104.0 | tail -1 | cut -d " " -f1`

otel-stop:
  docker stop "{{otel-container-id}}"

# Run `ftl dev` with the given args after setting the necessary envar.
otel-dev *args:
  #!/bin/bash
  set -euo pipefail

  export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:${OTEL_GRPC_PORT}"
  export OTEL_METRIC_EXPORT_INTERVAL=${OTEL_METRIC_EXPORT_INTERVAL}
  # Uncomment this line for much richer debug logs
  # export FTL_O11Y_LOG_LEVEL="debug"
  ftl dev {{args}}

# Runs a Grafana stack for storing and visualizing telemetry. This stack includes a
# Prometheus database for metrics and a Tempo database for traces; both of which are
# populated by the OTLP over GRPC collector that is integrated with this stack.
#
# Running `just otel-dev` will export ftl metrics to this Grafana stack.
grafana:
  docker compose up -d grafana

grafana-stop:
  docker compose down grafana

# Start storybook server
storybook:
  #!/bin/bash
  cd frontend && pnpm run storybook

# Build an FTL Docker image.
build-docker name:
  docker build --platform linux/amd64 -t ftl0/ftl-{{name}}:"${GITHUB_SHA:-$(git rev-parse HEAD)}" -t ftl0/ftl-{{name}}:latest -f Dockerfile.{{name}} .
