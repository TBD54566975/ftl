package observability

var (
	Calls      *CallMetrics
	Deployment *DeploymentMetrics
	Controller *ControllerTracing
)

func init() {
	Calls = initCallMetrics()
	Deployment = initDeploymentMetrics()
	Controller = initControllerTracing()
}
