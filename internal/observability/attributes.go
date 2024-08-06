package observability

const (
	ModuleNameAttribute          = "ftl.module.name"
	OutcomeStatusNameAttribute   = "ftl.outcome.status"
	RunnerDeploymentKeyAttribute = "ftl.deployment.key"

	SuccessStatus = "success"
	FailureStatus = "failure"
)

func SuccessOrFailureStatus(succeeded bool) string {
	if succeeded {
		return SuccessStatus
	}
	return FailureStatus
}
