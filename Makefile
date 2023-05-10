VERSION = $(shell git describe --tags --always --dirty)

COMMON_LOG_IN = common/log/api.go
COMMON_LOG_OUT = common/log/log_level_string.go

SCHEMA_IN = schema/schema.go schema/protobuf.go
SCHEMA_OUT = protos/xyz/block/ftl/v1/schema/schema.proto

SQLC_IN = sqlc.yaml \
		  backplane/internal/sql/schema/*.sql \
		  backplane/internal/sql/queries.sql
SQLC_OUT = backplane/internal/sql/db.go \
		   $(shell grep -q copyfrom backplane/internal/sql/queries.sql && echo backplane/internal/sql/copyfrom.go) \
		   backplane/internal/sql/models.go \
		   backplane/internal/sql/queries.sql.go

PROTO_IN = protos/buf.yaml \
		   protos/buf.gen.yaml \
		   protos/xyz/block/ftl/v1/ftl.proto \
		   protos/xyz/block/ftl/v1/schema/schema.proto
PROTO_OUT = protos/xyz/block/ftl/v1/ftlv1connect/ftl.connect.go \
			protos/xyz/block/ftl/v1/schema/schema.pb.go \
			protos/xyz/block/ftl/v1/ftl.pb.go \
			console/src/protos/xyz/block/ftl/v1/ftl_connect.ts \
			console/src/protos/xyz/block/ftl/v1/schema/schema_pb.ts \
			console/src/protos/xyz/block/ftl/v1/schema/runtime_pb.ts \
			console/src/protos/xyz/block/ftl/v1/ftl_pb.ts


.DEFAULT_GOAL := help

.PHONY: help
help: ## This help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY:
release:
	cd console && npm run build
	rm -rf build
	mkdir -p build
	go build -o build/ftl -tags release -ldflags "-X main.version=$(VERSION)" ./cmd/ftl 

.PHONY: generate
generate: $(SQLC_OUT) $(SCHEMA_OUT) $(PROTO_OUT) $(COMMON_LOG_OUT) ## Regenerate source.

.PHONY: protosync
protosync: ## Synchronise external protos into FTL repo.
	protosync

$(PROTO_OUT) &: $(PROTO_IN)
	buf format -w
	buf lint
	(cd protos && buf generate)

$(SCHEMA_OUT) &: $(SCHEMA_IN)
	ftl schema protobuf > $(SCHEMA_OUT)~ && mv $(SCHEMA_OUT)~ $(SCHEMA_OUT)

$(SQLC_OUT) &: $(SQLC_IN)
	sqlc generate --experimental

$(COMMON_LOG_OUT) &: $(COMMON_LOG_IN)
	go generate $<
