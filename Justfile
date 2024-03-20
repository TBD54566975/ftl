set positional-arguments

RELEASE := "build/release"
VERSION := `git describe --tags --always --dirty | sed -e 's/^v//'`
KT_RUNTIME_OUT := "kotlin-runtime/ftl-runtime/target/ftl-runtime-1.0-SNAPSHOT.jar"
KT_RUNTIME_RUNNER_TEMPLATE_OUT := "build/template/ftl/jars/ftl-runtime.jar"
RUNNER_TEMPLATE_ZIP := "backend/controller/scaling/localscaling/template.zip"
TIMESTAMP := `date +%s`
SCHEMA_OUT := "backend/protos/xyz/block/ftl/v1/schema/schema.proto"
ZIP_DIRS := "go-runtime/compile/build-template go-runtime/compile/external-module-template go-runtime/scaffolding kotlin-runtime/scaffolding kotlin-runtime/external-module-template"
FRONTEND_OUT := "frontend/dist/index.html"

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

# Build everything
build-all: build-frontend build-generate build-kt-runtime build-protos build-sqlc build-zips
  @just build ftl ftl-controller ftl-runner ftl-initdb

# Run "go generate" on all packages
build-generate:
  @mk backend/schema/aliaskind_enumer.go : backend/schema/metadataalias.go -- go generate -x ./backend/schema
  @mk internal/log/log_level_string.go : internal/log/api.go -- go generate -x ./internal/log

# Build command-line tools
build +tools: build-protos build-sqlc build-zips build-frontend
  #!/bin/bash
  shopt -s extglob
  for tool in $@; do mk "{{RELEASE}}/$tool" : !(build) -- go build -o "{{RELEASE}}/$tool" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./cmd/$tool"; done

export DATABASE_URL := "postgres://postgres:secret@localhost:54320/ftl?sslmode=disable"

# Explicitly initialise the database
init-db:
  dbmate drop || true
  dbmate create
  dbmate --migrations-dir backend/controller/sql/schema up

# Regenerate SQLC code
build-sqlc:
  @mk backend/controller/sql/{db.go,models.go,querier.go,queries.sql.go} : backend/controller/sql/queries.sql backend/controller/sql/schema -- sqlc generate --experimental

# Build the ZIP files that are embedded in the FTL release binaries
build-zips: build-kt-runtime
  @for dir in {{ZIP_DIRS}}; do (cd $dir && mk ../$(basename ${dir}).zip : . -- "rm -f $(basename ${dir}.zip) && zip -q --symlinks -r ../$(basename ${dir}).zip ."); done

# Rebuild frontend
build-frontend: npm-install
  @mk {{FRONTEND_OUT}} : frontend/src -- "cd frontend && npm run build"

# Build the Kotlin runtime (if necessary)
build-kt-runtime:
  @mk {{KT_RUNTIME_OUT}} : kotlin-runtime/ftl-runtime -- mvn -f kotlin-runtime/ftl-runtime -Dmaven.test.skip=true -B install
  @mk {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}} : {{KT_RUNTIME_OUT}} -- "mkdir -p $(dirname {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}}) && install -m 0600 {{KT_RUNTIME_OUT}} {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}}"
  @mk {{RUNNER_TEMPLATE_ZIP}} : {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}} -- "cd build/template && zip -q --symlinks -r ../../{{RUNNER_TEMPLATE_ZIP}} ."

# Install Node dependencies
npm-install:
  @mk frontend/node_modules : frontend/package.json frontend/src -- "cd frontend && npm install"

# Regenerate protos
build-protos: npm-install
  @mk {{SCHEMA_OUT}} : backend/schema -- "ftl-schema > {{SCHEMA_OUT}} && buf format -w && buf lint && cd backend/protos && buf generate"
