package observability

import (
	"go.opentelemetry.io/otel/attribute"
)

type metricAttributeBuilders struct {
	moduleName      func(name string) attribute.KeyValue
	featureName     func(name string) attribute.KeyValue
	destinationVerb func(name string) attribute.KeyValue
}

var metricAttributes = metricAttributeBuilders{
	moduleName: func(name string) attribute.KeyValue {
		return attribute.String("ftl.module.name", name)
	},
	featureName: func(name string) attribute.KeyValue {
		return attribute.String("ftl.feature.name", name)
	},
	destinationVerb: func(name string) attribute.KeyValue {
		return attribute.String("ftl.verb.dest", name)
	},
}
