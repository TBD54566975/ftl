package metrics

import (
	"github.com/TBD54566975/ftl/backend/schema"
	"go.opentelemetry.io/otel/attribute"
)

// FeatureNameAttribute identifies the feature (e.g. verb) that the associated
// metric originates from. The entire module qualified name is used:
// e.g. {module.verb}
func FeatureNameAttribute(ref schema.Ref) attribute.KeyValue {
	return attribute.String("ftl.feature.name", ref.Name)
}

// DestinationVerbAttribute identifies the target verb associated metric. This
// attribute is relevant for metrics involving verb invocations. The entire
// module qualified name is used: e.g. {module.verb}
func DestinationVerbAttribute(ref schema.Ref) attribute.KeyValue {
	return attribute.String("ftl.dest.verb", ref.Name)
}
