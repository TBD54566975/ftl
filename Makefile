VERSION = $(shell git describe --tags --always --dirty)

COMMON_LOG_IN = backend/common/log/api.go
COMMON_LOG_OUT = backend/common/log/log_level_string.go

SCHEMA_IN = backend/schema/schema.go backend/schema/protobuf.go cmd/ftl/cmd_schema.go
SCHEMA_OUT = protos/xyz/block/ftl/v1/schema/schema.proto

SQLC_IN = sqlc.yaml \
		  backend/controller/internal/sql/schema/*.sql \
		  backend/controller/internal/sql/queries.sql
SQLC_OUT = backend/controller/internal/sql/db.go \
		   $(shell grep -q copyfrom backend/controller/internal/sql/queries.sql && echo backend/controller/internal/sql/copyfrom.go) \
		   backend/controller/internal/sql/models.go \
		   backend/controller/internal/sql/queries.sql.go

PROTO_IN = protos/buf.yaml \
		   protos/buf.gen.yaml \
		   protos/xyz/block/ftl/v1/ftl.proto \
		   protos/xyz/block/ftl/v1/console/console.proto \
		   protos/xyz/block/ftl/v1/schema/schema.proto \
		   protos/xyz/block/ftl/v1/schema/runtime.proto
PROTO_OUT = protos/xyz/block/ftl/v1/ftlv1connect/ftl.connect.go \
			protos/xyz/block/ftl/v1/schema/schema.pb.go \
			protos/xyz/block/ftl/v1/console/console.pb.go \
			protos/xyz/block/ftl/v1/schema/runtime.pb.go \
			protos/xyz/block/ftl/v1/ftl.pb.go \
			console/client/src/protos/xyz/block/ftl/v1/ftl_connect.ts \
			console/client/src/protos/xyz/block/ftl/v1/schema/schema_pb.ts \
			console/client/src/protos/xyz/block/ftl/v1/schema/runtime_pb.ts \
			console/client/src/protos/xyz/block/ftl/v1/ftl_pb.ts \
			console/client/src/protos/xyz/block/ftl/v1/console/console_pb.ts
RELEASE_OUT = build/release/ftl build/release/ftl-controller build/release/ftl-runner

KT_RUNTIME_IN = $(shell find kotlin-runtime/ftl-runtime/src -name '*.kt')
KT_MVN_OUT = kotlin-runtime/ftl-runtime/target/ftl-runtime-1.0-SNAPSHOT-jar-with-dependencies.jar
KT_RUNTIME_OUT = build/template/ftl/jars/ftl-runtime.jar

.DEFAULT_GOAL := help

.PHONY: help
help: ## This help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: generate release ## Generate source and build binaries.

.PHONY: clean
clean: ## Clean build artifacts.
	rm -rf build $(SQLC_OUT) $(SCHEMA_OUT) $(PROTO_OUT) $(COMMON_LOG_OUT) $(RELEASE_OUT)
	mvn clean

.PHONY: release
release: build/release/ftl-controller build/release/ftl-runner build/release/ftl ## Build release binaries.

build/release/%: console/client/dist/index.html
	go build -o $@ -tags release -ldflags "-X main.version=$(VERSION) -X main.timestamp=$(shell date +%s)" ./cmd/$(shell basename $@)

$(KT_MVN_OUT): $(KT_RUNTIME_IN)
	mvn -pl :ftl-runtime clean package

$(KT_RUNTIME_OUT): $(KT_MVN_OUT)
	mkdir -p build/template/ftl/jars
	cp $< $@

console/client/dist/index.html:
	cd console/client && npm install && npm run build

.PHONY: generate
generate: $(PROTO_OUT) $(COMMON_LOG_OUT) $(SQLC_OUT) $(SCHEMA_OUT) ## Regenerate source.

.PHONY:
docker-runner: ## Build ftl-runner docker images.
	docker build --tag ghcr.io/tbd54566975/ftl-runner:latest --platform=linux/amd64 \
		-f Dockerfile.runner .

.PHONY:
docker-controller: ## Build ftl-controller docker images.
	docker build --tag ghcr.io/tbd54566975/ftl-controller:latest --platform=linux/amd64 \
		-f Dockerfile.controller .

.PHONY: protosync
protosync: ## Synchronise external protos into FTL repo.
	protosync

$(PROTO_OUT) &: $(PROTO_IN)
	buf format -w
	buf lint
	(cd protos && buf generate)
	(cd backend/common/3rdparty/protos && buf generate)

$(SCHEMA_OUT) &: $(SCHEMA_IN)
	ftl schema protobuf > $(SCHEMA_OUT)~ && mv $(SCHEMA_OUT)~ $(SCHEMA_OUT)

$(SQLC_OUT) &: $(SQLC_IN)
	sqlc generate --experimental
	# sqlc 1.18.0 generates a file with a missing import
	gosimports -w backend/controller/internal/sql/querier.go 

$(COMMON_LOG_OUT) &: $(COMMON_LOG_IN)
	go generate $<
