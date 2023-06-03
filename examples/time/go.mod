module github.com/TBD54566975/ftl/examples/time

go 1.20

replace github.com/TBD54566975/ftl => ../..

require go.opentelemetry.io/otel v1.16.0

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0 // indirect
)
