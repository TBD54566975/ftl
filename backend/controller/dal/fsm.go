package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	"github.com/TBD54566975/ftl/backend/controller/sql"
	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/backend/schema"
)

// StartFSMTransition sends an event to an executing instance of an FSM.
//
// If the instance doesn't exist a new one will be created.
//
// [name] is the name of the state machine to execute, [executionKey] is the
// unique identifier for this execution of the FSM.
//
// Returns ErrConflict if the state machine is already executing a transition.
//
// Note: this does not actually call the FSM, it just enqueues an async call for
// future execution.
//
// Note: no validation of the FSM is performed.
func (d *DAL) StartFSMTransition(ctx context.Context, fsm schema.RefKey, executionKey string, destinationState schema.RefKey, request json.RawMessage, retryParams schema.RetryParams) (err error) {
	encryptedRequest, err := d.encryptors.Async.EncryptJSON(request)
	if err != nil {
		return fmt.Errorf("failed to encrypt FSM request: %w", err)
	}

	// Create an async call for the event.
	origin := AsyncOriginFSM{FSM: fsm, Key: executionKey}
	asyncCallID, err := d.db.CreateAsyncCall(ctx, sql.CreateAsyncCallParams{
		Verb:              destinationState,
		Origin:            origin.String(),
		Request:           encryptedRequest,
		RemainingAttempts: int32(retryParams.Count),
		Backoff:           retryParams.MinBackoff,
		MaxBackoff:        retryParams.MaxBackoff,
	})
	observability.AsyncCalls.Created(ctx, destinationState, origin.String(), int64(retryParams.Count), err)
	if err != nil {
		return fmt.Errorf("failed to create FSM async call: %w", dalerrs.TranslatePGError(err))
	}
	queueDepth, err := d.db.AsyncCallQueueDepth(ctx)
	if err == nil {
		// Don't error out of an FSM transition just over a queue depth retrieval
		// error because this is only used for an observability gauge.
		observability.AsyncCalls.RecordQueueDepth(ctx, queueDepth)
	}

	// Start a transition.
	instance, err := d.db.StartFSMTransition(ctx, sql.StartFSMTransitionParams{
		Fsm:              fsm,
		Key:              executionKey,
		DestinationState: destinationState,
		AsyncCallID:      asyncCallID,
	})
	if err != nil {
		err = dalerrs.TranslatePGError(err)
		if errors.Is(err, dalerrs.ErrNotFound) {
			return fmt.Errorf("transition already executing: %w", dalerrs.ErrConflict)
		}
		return fmt.Errorf("failed to start FSM transition: %w", err)
	}
	if instance.CreatedAt.Equal(instance.UpdatedAt) {
		observability.FSM.InstanceCreated(ctx, fsm)
	}
	observability.FSM.TransitionStarted(ctx, fsm, destinationState)
	return nil
}

func (d *DAL) FinishFSMTransition(ctx context.Context, fsm schema.RefKey, instanceKey string) error {
	_, err := d.db.FinishFSMTransition(ctx, fsm, instanceKey)
	observability.FSM.TransitionCompleted(ctx, fsm)

	return dalerrs.TranslatePGError(err)
}

func (d *DAL) FailFSMInstance(ctx context.Context, fsm schema.RefKey, instanceKey string) error {
	_, err := d.db.FailFSMInstance(ctx, fsm, instanceKey)
	observability.FSM.InstanceCompleted(ctx, fsm)
	return dalerrs.TranslatePGError(err)
}

func (d *DAL) SucceedFSMInstance(ctx context.Context, fsm schema.RefKey, instanceKey string) error {
	_, err := d.db.SucceedFSMInstance(ctx, fsm, instanceKey)
	observability.FSM.InstanceCompleted(ctx, fsm)
	return dalerrs.TranslatePGError(err)
}

type FSMStatus = sql.FsmStatus

const (
	FSMStatusRunning   = sql.FsmStatusRunning
	FSMStatusCompleted = sql.FsmStatusCompleted
	FSMStatusFailed    = sql.FsmStatusFailed
)

type FSMInstance struct {
	leases.Lease
	// The FSM that this instance is executing.
	FSM schema.RefKey
	// The unique key for this instance.
	Key              string
	Status           FSMStatus
	CurrentState     optional.Option[schema.RefKey]
	DestinationState optional.Option[schema.RefKey]
}

// AcquireFSMInstance returns an FSM instance, also acquiring a lease on it.
//
// The lease must be released by the caller.
func (d *DAL) AcquireFSMInstance(ctx context.Context, fsm schema.RefKey, instanceKey string) (*FSMInstance, error) {
	lease, _, err := d.AcquireLease(ctx, leases.SystemKey("fsm_instance", fsm.String(), instanceKey), time.Second*5, optional.None[any]())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire FSM lease: %w", err)
	}
	row, err := d.db.GetFSMInstance(ctx, fsm, instanceKey)
	if err != nil {
		err = dalerrs.TranslatePGError(err)
		if !errors.Is(err, dalerrs.ErrNotFound) {
			return nil, err
		}
		row.Status = sql.FsmStatusRunning
	}
	return &FSMInstance{
		Lease:            lease,
		FSM:              fsm,
		Key:              instanceKey,
		Status:           row.Status,
		CurrentState:     row.CurrentState,
		DestinationState: row.DestinationState,
	}, nil
}
