set positional-arguments
set shell := ["bash", "-c"]

WATCHEXEC_ARGS := "-d 1s -e proto -e go -e sql -f sqlc.yaml --ignore **/types.ftl.go"
RELEASE := "build/release"
VERSION := `git describe --tags --always | sed -e 's/^v//'`
TIMESTAMP := `date +%s`
SCHEMA_OUT := "backend/protos/xyz/block/ftl/v1/schema/schema.proto"
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
PROTOS_OUT := "backend/protos/xyz/block/ftl/v1/console/console.pb.go " + \
              "backend/protos/xyz/block/ftl/v1/ftl.pb.go " + \
              "backend/protos/xyz/block/ftl/v1/schema/schema.pb.go " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/v1/console/console_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/v1/ftl_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/v1/schema/runtime_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/v1/schema/schema_pb.ts"

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
build-all: build-protos-unconditionally build-backend build-frontend build-generate build-sqlc build-zips lsp-generate build-jvm build-language-plugins

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
  mkdir -p frontend/console/dist
  touch frontend/console/dist/.phoney
  shopt -s extglob

  export CGO_ENABLED=0

  for tool in $@; do
    path="cmd/$tool"
    test "$tool" = "ftl" && path="frontend/cli"
    if [ "${FTL_DEBUG:-}" = "true" ]; then
      go build -o "{{RELEASE}}/$tool" -tags release -gcflags=all="-N -l" -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./$path"
    else
      mk "{{RELEASE}}/$tool" : !(build|integration|infrastructure|node_modules|Procfile*|Dockerfile*) -- go build -o "{{RELEASE}}/$tool" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./$path"
    fi
  done

# Build all backend binaries
build-backend:
  just build ftl ftl-controller ftl-runner

# Build all backend tests
build-backend-tests:
  go test -run ^NONE -tags integration,infrastructure ./...


build-jvm *args:
  mvn -f jvm-runtime/ftl-runtime install {{args}}

# Builds all language plugins
build-language-plugins: build-go build-python build-java build-kotlin

# Build ftl-language-go
build-go: build-zips build-protos
  #!/bin/bash
  shopt -s extglob

  if [ "${FTL_DEBUG:-}" = "true" ]; then
    go build -o "{{RELEASE}}/ftl-language-go" -tags release -gcflags=all="-N -l" -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./go-runtime/cmd/ftl-language-go"
  else
    mk "{{RELEASE}}/ftl-language-go" : !(build|integration) -- go build -o "{{RELEASE}}/ftl-language-go" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./go-runtime/cmd/ftl-language-go"
  fi

# Build ftl-language-python
build-python: build-zips build-protos
  #!/bin/bash
  shopt -s extglob

  if [ "${FTL_DEBUG:-}" = "true" ]; then
    go build -o "{{RELEASE}}/ftl-language-python" -tags release -gcflags=all="-N -l" -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./python-runtime/cmd/ftl-language-python"
  else
    mk "{{RELEASE}}/ftl-language-python" : !(build|integration) -- go build -o "{{RELEASE}}/ftl-language-python" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./python-runtime/cmd/ftl-language-python"
  fi

build-kotlin *args: build-zips build-protos
  #!/bin/bash
  shopt -s extglob
  if [ "${FTL_DEBUG:-}" = "true" ]; then
    go build -o "{{RELEASE}}/ftl-language-kotlin" -tags release -gcflags=all="-N -l" -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./go-runtime/cmd/ftl-language-kotlin"
  else
    mk "{{RELEASE}}/ftl-language-kotlin" : !(build|integration) -- go build -o "{{RELEASE}}/ftl-language-kotlin" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./jvm-runtime/cmd/ftl-language-kotlin"
  fi

build-java *args: build-zips build-protos
  #!/bin/bash
  shopt -s extglob

  if [ "${FTL_DEBUG:-}" = "true" ]; then
    go build -o "{{RELEASE}}/ftl-language-java" -tags release -gcflags=all="-N -l" -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./jvm-runtime/cmd/ftl-language-java"
  else
    mk "{{RELEASE}}/ftl-language-java" : !(build|integration) -- go build -o "{{RELEASE}}/ftl-language-java" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./jvm-runtime/cmd/ftl-language-java"
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
    -O 'go_package="github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema;schemapb"' \
    -O 'java_multiple_files=true' \
    xyz.block.ftl.v1.schema ./internal/schema.Schema && buf format -w && buf lint

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

localstack:
    docker compose up localstack -d --wait

localstack-stop:
    docker compose down localstack

# Start storybook server
storybook:
  @cd {{CONSOLE_ROOT}} && pnpm run storybook

# Build an FTL Docker image.
build-docker name:
  docker build \
    --platform linux/amd64 \
    -t ftl0/ftl-{{name}}:"${GITHUB_SHA:-$(git rev-parse HEAD)}" \
    -t ftl0/ftl-{{name}}:latest \
    -f Dockerfile.{{name}} .

chart *args:
    @cd charts && just {{args}}
