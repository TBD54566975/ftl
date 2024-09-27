// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sql

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
)

type Querier interface {
	// Reserve a pending async call for execution, returning the associated lease
	// reservation key and accompanying metadata.
	AcquireAsyncCall(ctx context.Context, ttl sqltypes.Duration) (AcquireAsyncCallRow, error)
	AssociateArtefactWithDeployment(ctx context.Context, arg AssociateArtefactWithDeploymentParams) error
	AsyncCallQueueDepth(ctx context.Context) (int64, error)
	BeginConsumingTopicEvent(ctx context.Context, subscription model.SubscriptionKey, event model.TopicEventKey) error
	CompleteEventForSubscription(ctx context.Context, name string, module string) error
	// Create a new artefact and return the artefact ID.
	CreateArtefact(ctx context.Context, digest []byte, content []byte) (int64, error)
	CreateAsyncCall(ctx context.Context, arg CreateAsyncCallParams) (int64, error)
	CreateCronJob(ctx context.Context, arg CreateCronJobParams) error
	CreateDeployment(ctx context.Context, moduleName string, schema []byte, key model.DeploymentKey) error
	CreateIngressRoute(ctx context.Context, arg CreateIngressRouteParams) error
	CreateRequest(ctx context.Context, origin Origin, key model.RequestKey, sourceAddr string) error
	DeleteSubscribers(ctx context.Context, deployment model.DeploymentKey) ([]model.SubscriberKey, error)
	DeleteSubscriptions(ctx context.Context, deployment model.DeploymentKey) ([]model.SubscriptionKey, error)
	DeregisterRunner(ctx context.Context, key model.RunnerKey) (int64, error)
	FailAsyncCall(ctx context.Context, error string, iD int64) (bool, error)
	FailAsyncCallWithRetry(ctx context.Context, arg FailAsyncCallWithRetryParams) (bool, error)
	FailFSMInstance(ctx context.Context, fsm schema.RefKey, key string) (bool, error)
	// Mark an FSM transition as completed, updating the current state and clearing the async call ID.
	FinishFSMTransition(ctx context.Context, fsm schema.RefKey, key string) (bool, error)
	GetActiveControllers(ctx context.Context) ([]Controller, error)
	GetActiveDeploymentSchemas(ctx context.Context) ([]GetActiveDeploymentSchemasRow, error)
	GetActiveDeployments(ctx context.Context) ([]GetActiveDeploymentsRow, error)
	GetActiveIngressRoutes(ctx context.Context) ([]GetActiveIngressRoutesRow, error)
	GetActiveRunners(ctx context.Context) ([]GetActiveRunnersRow, error)
	GetArtefactContentRange(ctx context.Context, start int32, count int32, iD int64) ([]byte, error)
	// Return the digests that exist in the database.
	GetArtefactDigests(ctx context.Context, digests [][]byte) ([]GetArtefactDigestsRow, error)
	GetCronJobByKey(ctx context.Context, key model.CronJobKey) (GetCronJobByKeyRow, error)
	GetDeployment(ctx context.Context, key model.DeploymentKey) (GetDeploymentRow, error)
	// Get all artefacts matching the given digests.
	GetDeploymentArtefacts(ctx context.Context, deploymentID int64) ([]GetDeploymentArtefactsRow, error)
	GetDeploymentsByID(ctx context.Context, ids []int64) ([]Deployment, error)
	// Get all deployments that have artefacts matching the given digests.
	GetDeploymentsWithArtefacts(ctx context.Context, digests [][]byte, schema []byte, count int64) ([]GetDeploymentsWithArtefactsRow, error)
	GetDeploymentsWithMinReplicas(ctx context.Context) ([]GetDeploymentsWithMinReplicasRow, error)
	GetExistingDeploymentForModule(ctx context.Context, name string) (GetExistingDeploymentForModuleRow, error)
	GetFSMInstance(ctx context.Context, fsm schema.RefKey, key string) (FsmInstance, error)
	// Get the runner endpoints corresponding to the given ingress route.
	GetIngressRoutes(ctx context.Context) ([]GetIngressRoutesRow, error)
	GetModulesByID(ctx context.Context, ids []int64) ([]Module, error)
	GetNextEventForSubscription(ctx context.Context, consumptionDelay sqltypes.Duration, topic model.TopicKey, cursor optional.Option[model.TopicEventKey]) (GetNextEventForSubscriptionRow, error)
	GetProcessList(ctx context.Context) ([]GetProcessListRow, error)
	GetRandomSubscriber(ctx context.Context, key model.SubscriptionKey) (GetRandomSubscriberRow, error)
	GetRunner(ctx context.Context, key model.RunnerKey) (GetRunnerRow, error)
	GetRunnersForDeployment(ctx context.Context, key model.DeploymentKey) ([]GetRunnersForDeploymentRow, error)
	GetSchemaForDeployment(ctx context.Context, key model.DeploymentKey) (*schema.Module, error)
	GetSubscription(ctx context.Context, column1 string, column2 string) (TopicSubscription, error)
	// Results may not be ready to be scheduled yet due to event consumption delay
	// Sorting ensures that brand new events (that may not be ready for consumption)
	// don't prevent older events from being consumed
	// We also make sure that the subscription belongs to a deployment that has at least one runner
	GetSubscriptionsNeedingUpdate(ctx context.Context) ([]GetSubscriptionsNeedingUpdateRow, error)
	GetTopic(ctx context.Context, dollar_1 int64) (Topic, error)
	GetTopicEvent(ctx context.Context, dollar_1 int64) (TopicEvent, error)
	GetUnscheduledCronJobs(ctx context.Context, startTime time.Time) ([]GetUnscheduledCronJobsRow, error)
	GetZombieAsyncCalls(ctx context.Context, limit int32) ([]AsyncCall, error)
	InsertSubscriber(ctx context.Context, arg InsertSubscriberParams) error
	InsertTimelineDeploymentCreatedEvent(ctx context.Context, arg InsertTimelineDeploymentCreatedEventParams) error
	InsertTimelineDeploymentUpdatedEvent(ctx context.Context, arg InsertTimelineDeploymentUpdatedEventParams) error
	IsCronJobPending(ctx context.Context, key model.CronJobKey, startTime time.Time) (bool, error)
	// Mark any controller entries that haven't been updated recently as dead.
	KillStaleControllers(ctx context.Context, timeout sqltypes.Duration) (int64, error)
	KillStaleRunners(ctx context.Context, timeout sqltypes.Duration) (int64, error)
	LoadAsyncCall(ctx context.Context, id int64) (AsyncCall, error)
	PopNextFSMEvent(ctx context.Context, fsm schema.RefKey, instanceKey string) (FsmNextEvent, error)
	PublishEventForTopic(ctx context.Context, arg PublishEventForTopicParams) error
	SetDeploymentDesiredReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int32) error
	SetNextFSMEvent(ctx context.Context, arg SetNextFSMEventParams) (int64, error)
	SetSubscriptionCursor(ctx context.Context, column1 model.SubscriptionKey, column2 model.TopicEventKey) error
	// Start a new FSM transition, populating the destination state and async call ID.
	//
	// "key" is the unique identifier for the FSM execution.
	StartFSMTransition(ctx context.Context, arg StartFSMTransitionParams) (FsmInstance, error)
	SucceedAsyncCall(ctx context.Context, response api.OptionalEncryptedAsyncColumn, iD int64) (bool, error)
	SucceedFSMInstance(ctx context.Context, fsm schema.RefKey, key string) (bool, error)
	UpdateCronJobExecution(ctx context.Context, arg UpdateCronJobExecutionParams) error
	UpsertController(ctx context.Context, key model.ControllerKey, endpoint string) (int64, error)
	UpsertModule(ctx context.Context, language string, name string) (int64, error)
	// Upsert a runner and return the deployment ID that it is assigned to, if any.
	UpsertRunner(ctx context.Context, arg UpsertRunnerParams) (int64, error)
	UpsertSubscription(ctx context.Context, arg UpsertSubscriptionParams) (UpsertSubscriptionRow, error)
	UpsertTopic(ctx context.Context, arg UpsertTopicParams) error
}

var _ Querier = (*Queries)(nil)
