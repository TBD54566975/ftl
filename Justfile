set positional-arguments
set shell := ["bash", "-c"]

WATCHEXEC_ARGS := "-d 1s -e proto -e go -e sql -f sqlc.yaml --ignore **/types.ftl.go"
RELEASE := "build/release"
VERSION := `git describe --tags --always | sed -e 's/^v//'`
TIMESTAMP := `date +%s`
SCHEMA_OUT := "backend/protos/xyz/block/ftl/schema/v1/schema.proto"
ZIP_DIRS := "go-runtime/compile/build-template " + \
            "go-runtime/compile/external-module-template " + \
            "go-runtime/compile/main-work-template " + \
            "internal/projectinit/scaffolding " + \
            "go-runtime/scaffolding " + \
            "jvm-runtime/java/scaffolding " + \
            "jvm-runtime/kotlin/scaffolding " + \
            "python-runtime/compile/build-template " + \
            "python-runtime/compile/external-module-template " + \
            "python-runtime/scaffolding"
CONSOLE_ROOT := "frontend/console"
FRONTEND_OUT := CONSOLE_ROOT + "/dist/index.html"
EXTENSION_OUT := "frontend/vscode/dist/extension.js"
PROTOS_IN := "backend/protos"
PROTOS_OUT := "backend/protos/xyz/block/ftl/console/v1/console.pb.go " + \
              "backend/protos/xyz/block/ftl//v1/ftl.pb.go " + \
              "backend/protos/xyz/block/ftl/schema/v1/schema.pb.go " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/console/v1/console_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/v1/ftl_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/schema/v1/schema_pb.ts" + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/publish/v1/publish_pb.ts"
# Configuration for building Docker images
DOCKER_IMAGES := '''
{
  "controller": {
    "extra_binaries": ["ftl"],
    "extra_files": ["ftl-provisioner-config.toml"]
  },
  "provisioner": {
    "extra_binaries": ["ftl-provisioner-cloudformation"],
    "extra_files": ["ftl-provisioner-config.toml"]
  },
  "cron": {},
  "http-ingress": {},
  "runner": {},
  "runner-jvm": {}
}
'''

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
  rm -rf frontend/console/dist
  rm -rf frontend/console/node_modules
  rm -rf python-runtime/ftl/.venv
  find . -name '*.zip' -exec rm {} \;
  mvn -f jvm-runtime/ftl-runtime clean

# Live rebuild the ftl binary whenever source changes.
live-rebuild:
  watchexec {{WATCHEXEC_ARGS}} -- "just build-sqlc && just build ftl"

# Run "ftl dev" with live-reloading whenever source changes.
dev *args:
  watchexec -r {{WATCHEXEC_ARGS}} -- "just build-sqlc && ftl dev --plain {{args}}"

# Build everything
build-all: build-protos-unconditionally build-backend build-frontend build-backend-tests build-generate build-sqlc build-zips lsp-generate build-jvm build-language-plugins

# Run "go generate" on all packages
build-generate:
  @mk internal/schema/aliaskind_enumer.go : internal/schema/metadataalias.go -- go generate -x ./internal/schema
  @mk internal/log/log_level_string.go : internal/log/api.go -- go generate -x ./internal/log

# Build command-line tools
build +tools: build-protos build-zips build-frontend
  @just build-without-frontend $@

# Build command-line tools
# This does not have a dependency on the frontend
# But it will be included if it was already built
build-without-frontend +tools: build-protos build-zips
  #!/bin/bash
  set -euo pipefail
  mkdir -p frontend/console/dist
  touch frontend/console/dist/.phoney
  shopt -s extglob

  export CGO_ENABLED=0

  for tool in $@; do
    path="cmd/$tool"
    test "$tool" = "ftl" && path="frontend/cli"
    just build-go-binary "./$path" "$tool"
  done

# Build all backend binaries
build-backend:
  just build ftl ftl-controller ftl-runner

# Build all backend tests
build-backend-tests:
  go test -run ^NONE -tags integration,infrastructure ./... > /dev/null


build-jvm *args:
  mvn -f jvm-runtime/ftl-runtime install {{args}}

# Builds all language plugins
build-language-plugins:
  @just build-go-binary ./go-runtime/cmd/ftl-language-go
  @just build-go-binary ./python-runtime/cmd/ftl-language-python
  @just build-go-binary ./jvm-runtime/cmd/ftl-language-java
  @just build-go-binary ./jvm-runtime/cmd/ftl-language-kotlin

# Build a Go binary with the correct flags and place it in the release dir
build-go-binary dir binary="": build-zips build-protos build-frontend
  #!/bin/bash
  set -euo pipefail
  shopt -s extglob

  binary="${2:-$(basename "$1")}"

  if [ "${FTL_DEBUG:-}" = "true" ]; then
    go build -o "{{RELEASE}}/${binary}" -tags release -gcflags=all="-N -l" -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "$1"
  else
    mk "{{RELEASE}}/${binary}" : !(build|integration|infrastructure|node_modules|Procfile*|Dockerfile*) -- go build -o "{{RELEASE}}/${binary}" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "$1"
  fi

export DATABASE_URL := "postgres://postgres:secret@localhost:15432/ftl?sslmode=disable"

# Explicitly initialise the database
init-db:
  dbmate drop || true
  dbmate create
  dbmate --no-dump-schema --migrations-dir backend/controller/sql/schema up

# Regenerate SQLC code (requires init-db to be run first)
build-sqlc:
  @mk $(eval echo $(yq '.sql[].gen.go.out + "/{db.go,models.go,querier.go,queries.sql.go}"' sqlc.yaml)) \
    : sqlc.yaml $(yq '.sql[].queries[]' sqlc.yaml) \
    --  "just init-db && sqlc generate"

# Build the ZIP files that are embedded in the FTL release binaries
build-zips:
  @for dir in {{ZIP_DIRS}}; do (cd $dir && mk ../$(basename ${dir}).zip : . -- "rm -f $(basename ${dir}.zip) && zip -q --symlinks -r ../$(basename ${dir}).zip ."); done

# Rebuild frontend
build-frontend: pnpm-install
  @mk {{FRONTEND_OUT}} : {{CONSOLE_ROOT}}/package.json {{CONSOLE_ROOT}}/src -- "cd {{CONSOLE_ROOT}} && pnpm run build"

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

# Build the IntelliJ plugin
build-intellij-plugin:
  @cd frontend/intellij && gradle buildPlugin

# Format console code.
format-frontend:
  cd {{CONSOLE_ROOT}} && pnpm run lint:fix

# Install Node dependencies using pnpm
pnpm-install:
  @for i in {1..3}; do mk frontend/**/node_modules : frontend/**/package.json -- "pnpm install --frozen-lockfile" && break || sleep 5; done

# Regenerate protos
build-protos:
  @mk {{SCHEMA_OUT}} : internal/schema -- "@just go2proto"
  @mk {{PROTOS_OUT}} : {{PROTOS_IN}} -- "@just build-protos-unconditionally"

# Generate .proto files from .go types.
go2proto:
  @mk "{{SCHEMA_OUT}}" : cmd/go2proto internal/schema -- go2proto -o "{{SCHEMA_OUT}}" \
    -O 'go_package="github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1;schemapb"' \
    -O 'java_multiple_files=true' \
    xyz.block.ftl.schema.v1 ./internal/schema.Schema && buf format -w && buf lint

# Unconditionally rebuild protos
build-protos-unconditionally: lint-protos pnpm-install go2proto
  cd backend/protos && buf generate

# Run integration test(s)
integration-tests *test:
  #!/bin/bash
  set -euo pipefail
  testName=${1:-}
  for i in {1..3}; do go test -fullpath -count 1 -v -tags integration -run "$testName" -p 1 $(find . -type f -name '*_test.go' -print0 | xargs -0 grep -r -l "$testName" | xargs grep -l '//go:build integration' | xargs -I {} dirname './{}') && break || true; done

# Run integration test(s)
infrastructure-tests *test:
  #!/bin/bash
  set -euo pipefail
  testName=${1:-}
  for i in {1..3}; do go test -fullpath -count 1 -v -tags infrastructure -run "$testName" -p 1 $(find . -type f -name '*_test.go' -print0 | xargs -0 grep -r -l "$testName" | xargs grep -l '//go:build infrastructure' | xargs -I {} dirname './{}') && break || true; done

# Run README doc tests
test-readme *args:
  mdcode run {{args}} README.md -- bash test.sh

# Run "go mod tidy" on all packages including tests
tidy:
  git ls-files | grep go.mod | grep -v '{{{{' | xargs -n1 dirname | xargs -I {} sh -c 'cd {} && echo {} && go mod tidy'

# Check for changes in existing SQL migrations compared to main
ensure-frozen-migrations:
  @scripts/ensure-frozen-migrations

# Run backend tests
test-backend:
  @gotestsum --hide-summary skipped --format-hide-empty-pkg -- -short -fullpath ./...

# Run shell script tests
test-scripts:
  GIT_COMMITTER_NAME="CI" \
    GIT_COMMITTER_EMAIL="no-reply@tbd.email" \
    GIT_AUTHOR_NAME="CI" \
    GIT_AUTHOR_EMAIL="no-reply@tbd.email" \
    scripts/tests/test-ensure-frozen-migrations.sh

# Test the frontend
test-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && pnpm run test

# Run end-to-end tests on the frontend
e2e-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && npx playwright install --with-deps && pnpm run e2e

# Lint the frontend
lint-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && pnpm run lint && tsc

# Lint .proto files
lint-protos:
  @buf lint

# Lint the backend
lint-backend:
  @golangci-lint run --new-from-rev=$(git merge-base origin/main HEAD) ./...
  @lint-commit-or-rollback ./backend/...
  @go-check-sumtype ./...

# Lint shell scripts.
lint-scripts:
  #!/bin/bash
  set -euo pipefail
  shellcheck -f gcc -e SC2016 $(find scripts -type f -not -path scripts/tests) | to-annotation

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

# Bring up localstack
localstack:
    docker compose up localstack -d --wait

# Bring down localstack
localstack-stop:
    docker compose down localstack

# Start storybook server
storybook:
  @cd {{CONSOLE_ROOT}} && pnpm run storybook

# Build an FTL Docker image
build-docker name:
  #!/bin/bash
  set -euo pipefail

  rm -rf build/release

  config="$(echo '{{DOCKER_IMAGES}}' | jq -r '."{{name}}"')"
  if [ "$config" = "null" ]; then
    echo "FATAL: No configuration found for {{name}}, update DOCKER_IMAGES"
    exit 1
  fi

  # Determine if this is a runner variant
  if [[ "{{name}}" =~ ^runner-(.+)$ ]]; then
    runtime="${BASH_REMATCH[1]}"
    # Build base runner first
    just build-docker runner
    # Build the language-specific runner
    docker build \
      --platform linux/amd64 \
      -t ftl0/ftl-{{name}}:latest \
      -t ftl0/ftl-{{name}}:"${GITHUB_SHA:-$(git rev-parse HEAD)}" \
      -f Dockerfile.runner-${runtime} .
  else
    # First build the binary on host
    extra_binaries="$(echo "$config" | jq -r '.extra_binaries // [] | join(" ")')"
    GOARCH=amd64 GOOS=linux CGO_ENABLED=0 just build ftl-{{name}} ${extra_binaries}
    # The main binary in the container must be named "svc"
    (cd build/release && mv ftl-{{name}} svc)

    extra_files="$(echo "$config" | jq -r '.extra_files // [] | join(" ")')"
    for file in $extra_files; do
      echo "Copying $file to build/release"
      cp "$file" build/release
    done

    # Build regular service
    docker build \
      --platform linux/amd64 \
      -t ftl0/ftl-{{name}}:latest \
      -t ftl0/ftl-{{name}}:"${GITHUB_SHA:-$(git rev-parse HEAD)}" \
      --build-arg SERVICE={{name}} \
      --build-arg PORT=8891 \
      --build-arg RUNTIME=$([ "{{name}}" = "runner" ] && echo "ubuntu-runtime" || echo "scratch-runtime") \
      --build-arg EXTRA_FILES="$(echo "$config" | jq -r '((.extra_files // []) + (.extra_binaries // [])) | join(" ")')" \
      -f Dockerfile build/release
  fi

# Build all Docker images
build-all-docker:
  @for image in $(just list-docker-images); do just build-docker $image; done

# List available Docker images
list-docker-images:
  @echo '{{DOCKER_IMAGES}}' | jq -r 'keys | join(" ")'

# Run docker compose up with all docker compose files
compose-up:
  #!/bin/bash
  set -o pipefail
  docker_compose_files="
  -f docker-compose.yml
  -f internal/dev/docker-compose.grafana.yml
  -f internal/dev/docker-compose.mysql.yml
  -f internal/dev/docker-compose.postgres.yml
  -f internal/dev/docker-compose.redpanda.yml
  -f internal/dev/docker-compose.registry.yml"

  docker compose -p "ftl" $docker_compose_files up -d --wait
  status=$?
  if [ $status -ne 0 ] && [ -n "${CI-}" ]; then
    # CI fails regularly due to network issues. Retry once.
    echo "docker compose up failed, retrying in 3s..."
    sleep 3
    docker compose -p "ftl" $docker_compose_files up -d --wait
  fi


# Run a Just command in the Helm charts directory
chart *args:
  @cd charts && just {{args}}
