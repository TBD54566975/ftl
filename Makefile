VERSION = $(shell git describe --tags --always --dirty)

BINARIES=ftl ftl-control-plane ftl-runner-go

COMMON_LOG_IN = internal/log/api.go
COMMON_LOG_OUT = internal/log/log_level_string.go

SCHEMA_IN = schema/schema.go schema/protobuf.go cmd/ftl/cmd_schema.go
SCHEMA_OUT = protos/xyz/block/ftl/v1/schema/schema.proto

SQLC_IN = sqlc.yaml \
		  controlplane/internal/sql/schema/*.sql \
		  controlplane/internal/sql/queries.sql
SQLC_OUT = controlplane/internal/sql/db.go \
		   $(shell grep -q copyfrom controlplane/internal/sql/queries.sql && echo controlplane/internal/sql/copyfrom.go) \
		   controlplane/internal/sql/models.go \
		   controlplane/internal/sql/queries.sql.go

PROTO_IN = protos/buf.yaml \
		   protos/buf.gen.yaml \
		   protos/xyz/block/ftl/v1/ftl.proto \
		   protos/xyz/block/ftl/v1/console/console.proto \
		   protos/xyz/block/ftl/v1/schema/schema.proto \
		   internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/metrics_service.proto
PROTO_OUT = protos/xyz/block/ftl/v1/ftlv1connect/ftl.connect.go \
			protos/xyz/block/ftl/v1/schema/schema.pb.go \
			protos/xyz/block/ftl/v1/console/console.pb.go \
			protos/xyz/block/ftl/v1/ftl.pb.go \
			console/src/protos/xyz/block/ftl/v1/ftl_connect.ts \
			console/src/protos/xyz/block/ftl/v1/schema/schema_pb.ts \
			console/src/protos/xyz/block/ftl/v1/schema/runtime_pb.ts \
			console/src/protos/xyz/block/ftl/v1/ftl_pb.ts \
			console/src/protos/xyz/block/ftl/v1/console/console_pb.ts \
			internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/v1connect/metrics_service.connect.go


.DEFAULT_GOAL := help

.PHONY: help
help: ## This help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY:
release:
	cd console/client && npm run build
	rm -rf build
	mkdir -p build
	for binary in $(BINARIES); do \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/$$binary-linux-amd64 -tags release -ldflags "-X main.version=$(VERSION)" ./cmd/$$binary ; \
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/$$binary-darwin-amd64 -tags release -ldflags "-X main.version=$(VERSION)" ./cmd/$$binary ; \
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/$$binary-darwin-arm64 -tags release -ldflags "-X main.version=$(VERSION)" ./cmd/$$binary ; \
	done

.PHONY: generate
generate: $(SQLC_OUT) $(SCHEMA_OUT) $(PROTO_OUT) $(COMMON_LOG_OUT) ## Regenerate source.

.PHONY: protosync
protosync: ## Synchronise external protos into FTL repo.
	protosync

$(PROTO_OUT) &: $(PROTO_IN)
	buf format -w
	buf lint
	(cd protos && buf generate)
	(cd internal/3rdparty/protos && buf generate)

$(SCHEMA_OUT) &: $(SCHEMA_IN)
	ftl schema protobuf > $(SCHEMA_OUT)~ && mv $(SCHEMA_OUT)~ $(SCHEMA_OUT)

$(SQLC_OUT) &: $(SQLC_IN)
	sqlc generate --experimental
	# sqlc 1.18.0 generates a file with a missing import
	gosimports -w controlplane/internal/sql/querier.go 

$(COMMON_LOG_OUT) &: $(COMMON_LOG_IN)
	go generate $<
