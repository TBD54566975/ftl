.PHONY: help
help: ## This help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: dev
dev: ## Run hot reload dev server.
	reflex -d fancy -c reflex.conf  # https://github.com/cespare/reflex

.PHONY: protos
protos: ## Regenerate protos.
	buf lint
	(cd common/protos && buf generate)

.PHONY: protosync
protosync: ## Synchronise external protos into FTL repo.
	protosync
