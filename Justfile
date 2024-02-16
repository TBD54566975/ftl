# Start a hot-reloading dev cluster
dev: install-jars
  goreman -logtime=false start

# Install all JARs to local Maven repository and local build directory
install-jars:
  bit 'build/**/*.jar'

# Deploy the Go time module
deploy-time:
  ftl deploy examples/time

# Deploy the Kotlin echo module
deploy-echo-kotlin:
  ftl deploy examples/echo-kotlin

regen-schema:
  bit backend/protos/xyz/block/ftl/v1/schema/schema.proto
  bit backend/protos/xyz/block/ftl/v1/schema/schema.pb.go

# Run errtrace on Go files to add stacks
errtrace:
  git ls-files -z -- '*.go' | grep -zv /_ | xargs -0 errtrace -w && go mod tidy