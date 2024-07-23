package metrics

import (
	"github.com/TBD54566975/ftl/backend/schema"
	"go.opentelemetry.io/otel/attribute"
)

// ModuleNameAttribute identifies the name of the module that the associated
// metric originates from.
func ModuleNameAttribute(name string) attribute.KeyValue {
	return attribute.String("ftl.module.name", name)
}

// FeatureNameAttribute identifies the feature (e.g. verb) that the associated
// metric originates from. The entire module qualified name is used:
// e.g. {module.verb}
func FeatureNameAttribute(ref schema.Ref) attribute.KeyValue {
	return attribute.String("ftl.feature.ref", ref.Name)
}
