GEN_SCHEMA_PROTO = protos/xyz/block/ftl/v1/schema/schema.proto

.PHONY: help
help: ## This help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: dev
dev: ## Run hot reload dev server.
	reflex -d fancy -c reflex.conf  # https://github.com/cespare/reflex

.PHONY: generate
generate: ## Regenerate source.
	ftl schema protobuf > $(GEN_SCHEMA_PROTO)~ && mv $(GEN_SCHEMA_PROTO)~ $(GEN_SCHEMA_PROTO)
	buf format -w
	buf lint
	(cd protos && buf generate)
	go generate ./...

.PHONY: protosync
protosync: ## Synchronise external protos into FTL repo.
	protosync
