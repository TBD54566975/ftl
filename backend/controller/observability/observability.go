package observability

var (
	AsyncCalls *AsyncCallMetrics
	Calls      *CallMetrics
	Deployment *DeploymentMetrics
	PubSub     *PubSubMetrics
	Controller *ControllerTracing
)

func init() {
	AsyncCalls = initAsyncCallMetrics()
	Calls = initCallMetrics()
	Deployment = initDeploymentMetrics()
	PubSub = initPubSubMetrics()
	Controller = initControllerTracing()
}
