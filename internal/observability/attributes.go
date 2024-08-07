package observability

import "go.opentelemetry.io/otel/attribute"

const (
	ModuleNameAttribute          = "ftl.module.name"
	OutcomeStatusNameAttribute   = "ftl.outcome.status"
	RunnerDeploymentKeyAttribute = "ftl.deployment.key"

	SuccessStatus = "success"
	FailureStatus = "failure"
)

func SuccessOrFailureStatusAttr(succeeded bool) attribute.KeyValue {
	if succeeded {
		return attribute.String(OutcomeStatusNameAttribute, SuccessStatus)
	}
	return attribute.String(OutcomeStatusNameAttribute, FailureStatus)
}
