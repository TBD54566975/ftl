package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	"github.com/TBD54566975/ftl/backend/controller/pubsub/internal/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/backend/controller/state"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlpubsubv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/routing"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

const (
	// Events can be added simultaneously, which can cause events with out of order create_at values
	// By adding a delay, we ensure that by the time we read the events, no new events will be added
	// with earlier created_at values.
	eventConsumptionDelay = 100 * time.Millisecond
)

type Scheduler interface {
	Singleton(name string, retry backoff.Backoff, job scheduledtask.Job)
	Parallel(name string, retry backoff.Backoff, job scheduledtask.Job)
}

type Service struct {
	dal             *dal.DAL
	eventPublished  chan struct{}
	routeTable      *routing.RouteTable
	verbRouting     *routing.VerbCallRouter
	asyncCallsLock  sync.Mutex
	controllerState state.ControllerState
}

func New(ctx context.Context, conn libdal.Connection, rt *routing.RouteTable, controllerState state.ControllerState) *Service {
	m := &Service{
		dal:             dal.New(conn),
		eventPublished:  make(chan struct{}),
		routeTable:      rt,
		verbRouting:     routing.NewVerbRouterFromTable(ctx, rt),
		controllerState: controllerState,
	}
	go m.watchEventStream(ctx)
	go m.poll(ctx)
	return m
}

// poll waits for an event to be published (incl eventConsumptionDelay) or for a poll interval to pass
func (s *Service) poll(ctx context.Context) {
	logger := log.FromContext(ctx).Scope("pubsub")
	var publishedAt optional.Option[time.Time]
	for {
		var publishTrigger <-chan time.Time
		if pub, ok := publishedAt.Get(); ok {
			publishTrigger = time.After(time.Until(pub.Add(eventConsumptionDelay)))
		}

		// poll interval with jitter (1s - 1.1s)
		poll := time.Millisecond * (time.Duration(rand.Float64())*(100.0) + 1000.0) //nolint:gosec

		select {
		case <-ctx.Done():
			return

		case <-s.eventPublished:
			// published an event, so now we wait for eventConsumptionDelay before trying to progress subscriptions
			if !publishedAt.Ok() {
				publishedAt = optional.Some(time.Now())
			}

		case <-publishTrigger:
			// an event has been published and we have waited for eventConsumptionDelay
			if err := s.progressSubscriptions(ctx); err != nil {
				logger.Warnf("%s", err)
			}
			publishedAt = optional.None[time.Time]()

		case <-time.After(poll):
			if err := s.progressSubscriptions(ctx); err != nil {
				logger.Warnf("%s", err)
			}
		}
	}
}

func (s *Service) progressSubscriptions(ctx context.Context) error {
	count, err := s.dal.ProgressSubscriptions(ctx, eventConsumptionDelay)
	if err != nil {
		return fmt.Errorf("progress subscriptions: %w", err)
	}
	if count > 0 {
		// notify controller that we added an async call
		s.AsyncCallWasAdded(ctx)
	}
	return nil
}

func (s *Service) PublishEventForTopic(ctx context.Context, module, topic, caller string, payload []byte) error {
	err := s.dal.PublishEventForTopic(ctx, module, topic, caller, payload)
	if err != nil {
		return fmt.Errorf("%s.%s: publish: %w", module, topic, err)
	}
	s.eventPublished <- struct{}{}
	return nil
}

func (s *Service) resetSubscription(ctx context.Context, module, name string) (err error) {
	err = s.dal.ResetSubscription(ctx, module, name)
	if err != nil {
		return fmt.Errorf("%s.%s: reset: %w", module, name, err)
	}
	return nil
}

// OnCallCompletion is called within a transaction after an async call has completed to allow the subscription state to be updated.
func (s *Service) OnCallCompletion(ctx context.Context, tx libdal.Connection, origin async.AsyncOriginPubSub, failed bool, isFinalResult bool) error {
	if !isFinalResult {
		// Wait for the async call's retries to complete before progressing the subscription
		return nil
	}
	err := s.dal.Adopt(tx).CompleteEventForSubscription(ctx, origin.Subscription.Module, origin.Subscription.Name)
	if err != nil {
		return fmt.Errorf("%s: complete: %w", origin, err)
	}
	return nil
}

// AsyncCallDidCommit is called after a subscription's async call has been completed and committed to the database.
func (s *Service) AsyncCallDidCommit(ctx context.Context, origin async.AsyncOriginPubSub) {
	if err := s.progressSubscriptions(ctx); err != nil {
		log.FromContext(ctx).Scope("pubsub").Errorf(err, "failed to progress subscriptions")
	}
}

func (s *Service) CreateSubscriptions(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	err := s.dal.CreateSubscriptions(ctx, key, module)
	if err != nil {
		return fmt.Errorf("create subscriptions: %w", err)
	}
	return nil
}

func (s *Service) RemoveSubscriptionsAndSubscribers(ctx context.Context, key model.DeploymentKey) error {
	err := s.dal.RemoveSubscriptionsAndSubscribers(ctx, key)
	if err != nil {
		return fmt.Errorf("remove subscriptions and subscribers: %w", err)
	}
	return nil
}

func (s *Service) CreateSubscribers(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	err := s.dal.CreateSubscribers(ctx, key, module)
	if err != nil {
		return fmt.Errorf("create subscribers: %w", err)
	}
	return nil
}

func (s *Service) PublishEvent(ctx context.Context, req *connect.Request[ftlpubsubv1.PublishEventRequest]) (*connect.Response[ftlpubsubv1.PublishEventResponse], error) {
	// Publish the event.
	now := time.Now().UTC()
	pubishError := optional.None[string]()
	err := s.PublishEventForTopic(ctx, req.Msg.Topic.Module, req.Msg.Topic.Name, req.Msg.Caller, req.Msg.Body)
	if err != nil {
		pubishError = optional.Some(err.Error())
	}

	requestKey := optional.None[string]()
	if rk, err := rpc.RequestKeyFromContext(ctx); err == nil {
		if rk, ok := rk.Get(); ok {
			requestKey = optional.Some(rk.String())
		}
	}

	// Add to timeline.
	module := req.Msg.Topic.Module
	routes := s.routeTable.Current()
	route, ok := routes.GetDeployment(module).Get()
	if ok {
		timeline.ClientFromContext(ctx).Publish(ctx, timeline.PubSubPublish{
			DeploymentKey: route,
			RequestKey:    requestKey,
			Time:          now,
			SourceVerb:    schema.Ref{Name: req.Msg.Caller, Module: req.Msg.Topic.Module},
			Topic:         req.Msg.Topic.Name,
			Request:       req.Msg.Body,
			Error:         pubishError,
		})
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to publish a event to topic %s:%s: %w", req.Msg.Topic.Module, req.Msg.Topic.Name, err))
	}
	return connect.NewResponse(&ftlpubsubv1.PublishEventResponse{}), nil
}

func (s *Service) ResetSubscription(ctx context.Context, req *connect.Request[ftlpubsubv1.ResetSubscriptionRequest]) (*connect.Response[ftlpubsubv1.ResetSubscriptionResponse], error) {
	err := s.resetSubscription(ctx, req.Msg.Subscription.Module, req.Msg.Subscription.Name)
	if err != nil {
		return nil, fmt.Errorf("could not reset subscription: %w", err)
	}
	return connect.NewResponse(&ftlpubsubv1.ResetSubscriptionResponse{}), nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// AsyncCallWasAdded is an optional notification that an async call was added by this controller
//
// It allows us to speed up execution of scheduled async calls rather than waiting for the next poll time.
func (s *Service) AsyncCallWasAdded(ctx context.Context) {
	go func() {
		if _, err := s.ExecuteAsyncCalls(ctx); err != nil {
			log.FromContext(ctx).Errorf(err, "failed to progress subscriptions")
		}
	}()
}

func (s *Service) ExecuteAsyncCalls(ctx context.Context) (interval time.Duration, returnErr error) {
	// There are multiple entry points into this function, but we want the controller to handle async calls one at a time.
	s.asyncCallsLock.Lock()
	defer s.asyncCallsLock.Unlock()

	logger := log.FromContext(ctx)
	logger.Tracef("Acquiring async call")

	now := time.Now().UTC()
	sstate := s.routeTable.Current()

	enqueueTimelineEvent := func(call *dal.AsyncCall, err optional.Option[error]) {
		module := call.Verb.Module
		deployment, ok := sstate.GetDeployment(module).Get()
		if ok {
			eventType := timeline.AsyncExecuteEventTypeUnkown
			switch call.Origin.(type) {
			case async.AsyncOriginCron:
				eventType = timeline.AsyncExecuteEventTypeCron
			case async.AsyncOriginPubSub:
				eventType = timeline.AsyncExecuteEventTypePubSub
			case *async.AsyncOriginPubSub:
				eventType = timeline.AsyncExecuteEventTypePubSub
			default:
				break
			}
			errStr := optional.None[string]()
			if e, ok := err.Get(); ok {
				errStr = optional.Some(e.Error())
			}
			timeline.ClientFromContext(ctx).Publish(ctx, timeline.AsyncExecute{
				DeploymentKey: deployment,
				RequestKey:    call.ParentRequestKey,
				EventType:     eventType,
				Verb:          *call.Verb.ToRef(),
				Time:          now,
				Error:         errStr,
			})
		}
	}

	call, leaseCtx, err := s.dal.AcquireAsyncCall(ctx)
	if errors.Is(err, libdal.ErrNotFound) {
		logger.Tracef("No async calls to execute")
		return time.Second * 2, nil
	} else if err != nil {
		if call == nil {
			observability.AsyncCalls.AcquireFailed(ctx, err)
		} else {
			observability.AsyncCalls.Acquired(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, call.Catching, err)
			enqueueTimelineEvent(call, optional.Some(err))
		}
		return 0, fmt.Errorf("failed to acquire async call: %w", err)
	}
	// use originalCtx for things that should are done outside of the lease lifespan
	originalCtx := ctx
	ctx = leaseCtx

	// Extract the otel context from the call
	ctx, err = observability.ExtractTraceContextToContext(ctx, call.TraceContext)
	if err != nil {
		observability.AsyncCalls.Acquired(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, call.Catching, err)
		enqueueTimelineEvent(call, optional.Some(err))
		return 0, fmt.Errorf("failed to extract trace context: %w", err)
	}

	observability.AsyncCalls.Acquired(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, call.Catching, nil)

	defer func() {
		if returnErr == nil {
			// Post-commit notification based on origin
			switch origin := call.Origin.(type) {
			case async.AsyncOriginCron:
				break

			case async.AsyncOriginPubSub:
				go s.AsyncCallDidCommit(originalCtx, origin)

			default:
				break
			}
		}
	}()

	logger = logger.Scope(fmt.Sprintf("%s:%s", call.Origin, call.Verb)).Module(call.Verb.Module)

	if call.Catching {
		// Retries have been exhausted but catch verb has previously failed
		// We need to try again to catch the async call
		return 0, s.catchAsyncCall(ctx, logger, call)
	}

	logger.Tracef("Executing async call")
	req := &ftlv1.CallRequest{
		Verb:     call.Verb.ToProto(),
		Body:     call.Request,
		Metadata: metadataForAsyncCall(call),
	}
	resp, err := s.verbRouting.Call(ctx, connect.NewRequest(req))
	var callResult either.Either[[]byte, string]
	if err != nil {
		logger.Warnf("Async call could not be called: %v", err)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, optional.Some("async call could not be called"))
		callResult = either.RightOf[[]byte](err.Error())
	} else if perr := resp.Msg.GetError(); perr != nil {
		logger.Warnf("Async call failed: %s", perr.Message)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, optional.Some("async call failed"))
		callResult = either.RightOf[[]byte](perr.Message)
	} else {
		logger.Debugf("Async call succeeded")
		callResult = either.LeftOf[string](resp.Msg.GetBody())
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, optional.None[string]())
	}

	queueDepth := call.QueueDepth
	didScheduleAnotherCall, err := s.dal.CompleteAsyncCall(ctx, call, callResult, func(tx *dal.DAL, isFinalResult bool) error {
		return s.finaliseAsyncCall(ctx, tx, call, callResult, isFinalResult)
	})
	if err != nil {
		observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, queueDepth, err)
		enqueueTimelineEvent(call, optional.Some(err))
		return 0, fmt.Errorf("failed to complete async call: %w", err)
	}
	if !didScheduleAnotherCall {
		// Queue depth is queried at acquisition time, which means it includes the async
		// call that was just executed so we need to decrement
		queueDepth = call.QueueDepth - 1
	}
	observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, queueDepth, nil)
	enqueueTimelineEvent(call, optional.None[error]())
	return 0, nil
}

func (s *Service) catchAsyncCall(ctx context.Context, logger *log.Logger, call *dal.AsyncCall) error {
	catchVerb, ok := call.CatchVerb.Get()
	if !ok {
		logger.Warnf("Async call %s could not catch, missing catch verb", call.Verb)
		return fmt.Errorf("async call %s could not catch, missing catch verb", call.Verb)
	}
	logger.Debugf("Catching async call %s with %s", call.Verb, catchVerb)

	routeView := s.routeTable.Current()
	sch := routeView.Schema()

	verb := &schema.Verb{}
	if err := sch.ResolveToType(call.Verb.ToRef(), verb); err != nil {
		logger.Warnf("Async call %s could not catch, could not resolve original verb: %s", call.Verb, err)
		return fmt.Errorf("async call %s could not catch, could not resolve original verb: %w", call.Verb, err)
	}

	originalError := call.Error.Default("unknown error")
	originalResult := either.RightOf[[]byte](originalError)

	request := map[string]any{
		"verb": map[string]string{
			"module": call.Verb.Module,
			"name":   call.Verb.Name,
		},
		"requestType": verb.Request.String(),
		"request":     call.Request,
		"error":       originalError,
	}
	body, err := json.Marshal(request)
	if err != nil {
		logger.Warnf("Async call %s could not marshal body while catching", call.Verb)
		return fmt.Errorf("async call %s could not marshal body while catching", call.Verb)
	}

	req := &ftlv1.CallRequest{
		Verb:     catchVerb.ToProto(),
		Body:     body,
		Metadata: metadataForAsyncCall(call),
	}
	resp, err := s.verbRouting.Call(ctx, connect.NewRequest(req))
	var catchResult either.Either[[]byte, string]
	if err != nil {
		// Could not call catch verb
		logger.Warnf("Async call %s could not call catch verb %s: %s", call.Verb, catchVerb, err)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.Some("async call could not be called"))
		catchResult = either.RightOf[[]byte](err.Error())
	} else if perr := resp.Msg.GetError(); perr != nil {
		// Catch verb failed
		logger.Warnf("Async call %s had an error while catching (%s): %s", call.Verb, catchVerb, perr.Message)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.Some("async call failed"))
		catchResult = either.RightOf[[]byte](perr.Message)
	} else {
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.None[string]())
		catchResult = either.LeftOf[string](resp.Msg.GetBody())
	}
	queueDepth := call.QueueDepth
	didScheduleAnotherCall, err := s.dal.CompleteAsyncCall(ctx, call, catchResult, func(tx *dal.DAL, isFinalResult bool) error {
		// Exposes the original error to external components such as PubSub
		return s.finaliseAsyncCall(ctx, tx, call, originalResult, isFinalResult)
	})
	if err != nil {
		logger.Errorf(err, "Async call %s could not complete after catching (%s)", call.Verb, catchVerb)
		observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, queueDepth, err)
		return fmt.Errorf("async call %s could not complete after catching (%s): %w", call.Verb, catchVerb, err)
	}
	if !didScheduleAnotherCall {
		// Queue depth is queried at acquisition time, which means it includes the async
		// call that was just executed so we need to decrement
		queueDepth = call.QueueDepth - 1
	}
	logger.Debugf("Caught async call %s with %s", call.Verb, catchVerb)
	observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, queueDepth, nil)
	return nil
}

// ReapAsyncCalls fails async calls that have had their leases reaped
func (s *Service) ReapAsyncCalls(ctx context.Context) (nextInterval time.Duration, err error) {
	tx, err := s.dal.Begin(ctx)
	if err != nil {
		return 0, connect.NewError(connect.CodeInternal, fmt.Errorf("could not start transaction: %w", err))
	}
	defer tx.CommitOrRollback(ctx, &err)

	limit := 20
	calls, err := tx.GetZombieAsyncCalls(ctx, 20)
	if err != nil {
		return 0, fmt.Errorf("failed to get zombie async calls: %w", err)
	}
	for _, call := range calls {
		callResult := either.RightOf[[]byte]("async call lease expired")
		_, err := tx.CompleteAsyncCall(ctx, call, callResult, func(tx *dal.DAL, isFinalResult bool) error {
			return s.finaliseAsyncCall(ctx, tx, call, callResult, isFinalResult)
		})
		if err != nil {
			return 0, fmt.Errorf("failed to complete zombie async call: %w", err)
		}
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.Some("async call lease failed"))
	}

	if len(calls) == limit {
		return 0, nil
	}
	return time.Second * 5, nil
}

func metadataForAsyncCall(call *dal.AsyncCall) *ftlv1.Metadata {
	switch call.Origin.(type) {
	case async.AsyncOriginCron:
		return &ftlv1.Metadata{}

	case async.AsyncOriginPubSub:
		return &ftlv1.Metadata{}

	default:
		panic(fmt.Errorf("unsupported async call origin: %v", call.Origin))
	}
}

func (s *Service) finaliseAsyncCall(ctx context.Context, tx *dal.DAL, call *dal.AsyncCall, callResult either.Either[[]byte, string], isFinalResult bool) error {
	_, failed := callResult.(either.Right[[]byte, string])

	// Allow for handling of completion based on origin
	switch origin := call.Origin.(type) {
	case async.AsyncOriginPubSub:
		if err := s.OnCallCompletion(ctx, tx.Connection, origin, failed, isFinalResult); err != nil {
			return fmt.Errorf("failed to finalize pubsub async call: %w", err)
		}

	default:
		panic(fmt.Errorf("unsupported async call origin: %v", call.Origin))
	}
	return nil
}

func (s *Service) watchEventStream(ctx context.Context) {
	sub := s.controllerState.Subscribe(ctx)
	logger := log.FromContext(ctx).Scope("pubsub")
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-sub:
			switch e := event.(type) {
			case *state.DeploymentActivatedEvent:
				view := s.controllerState.View()
				deployment, err := view.GetDeployment(e.Key)
				if err != nil {
					logger.Errorf(err, "Deployment %s not found", e.Key)
					continue
				}
				err = s.CreateSubscriptions(ctx, e.Key, deployment.Schema)
				if err != nil {
					logger.Errorf(err, "Failed to create subscriptions for %s", e.Key)
					continue
				}
				err = s.CreateSubscribers(ctx, e.Key, deployment.Schema)
				if err != nil {
					logger.Errorf(err, "Failed to create subscribers for %s", e.Key)
					continue
				}
			case *state.DeploymentDeactivatedEvent:
				err := s.RemoveSubscriptionsAndSubscribers(ctx, e.Key)
				if err != nil {
					logger.Errorf(err, "Could not remove subscriptions and subscribers")
				}
			}
		}
	}
}
