package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/schema"
)

// SendFSMEvent sends an event to an executing instance of an FSM.
//
// If the instance doesn't exist a new one will be created.
//
// [name] is the name of the state machine to execute, [executionKey] is the
// unique identifier for this execution of the FSM.
//
// Returns ErrConflict if the state machine is already executing.
//
// Note: this does not actually call the FSM, it just enqueues an async call for
// future execution.
//
// Note: no validation of the FSM is performed.
func (d *DAL) SendFSMEvent(ctx context.Context, name, executionKey, destinationState string, verb schema.Ref, request json.RawMessage) error {
	_, err := d.db.SendFSMEvent(ctx, sql.SendFSMEventParams{
		Key:     executionKey,
		Name:    name,
		State:   destinationState,
		Verb:    verb,
		Request: request,
	})
	return translatePGError(err)
}

// AsyncCallOrigin represents the kind of originator of the async call.
type AsyncCallOrigin sql.AsyncCallOrigin

const (
	AsyncCallOriginFSM    = AsyncCallOrigin(sql.AsyncCallOriginFsm)
	AsyncCallOriginCron   = AsyncCallOrigin(sql.AsyncCallOriginCron)
	AsyncCallOriginPubSub = AsyncCallOrigin(sql.AsyncCallOriginPubsub)
)

type AsyncCall struct {
	*Lease
	ID     int64
	Origin AsyncCallOrigin
	// A key identifying the origin, e.g. the key of the FSM, cron job reference, etc.
	OriginKey string
	Verb      schema.Ref
	Request   json.RawMessage
}

// AcquireAsyncCall acquires a pending async call to execute.
//
// Returns ErrNotFound if there are no async calls to acquire.
func (d *DAL) AcquireAsyncCall(ctx context.Context) (*AsyncCall, error) {
	ttl := time.Second * 5
	row, err := d.db.AcquireAsyncCall(ctx, ttl)
	if err != nil {
		err = translatePGError(err)
		// We get a NULL constraint violation if there are no async calls to acquire, so translate it to ErrNotFound.
		if errors.Is(err, ErrConstraint) {
			return nil, fmt.Errorf("no pending async calls: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to acquire async call: %w", err)
	}
	return &AsyncCall{
		ID:        row.AsyncCallID,
		Verb:      row.Verb,
		Origin:    AsyncCallOrigin(row.Origin),
		OriginKey: row.OriginKey,
		Request:   row.Request,
		Lease:     d.newLease(ctx, row.LeaseKey, row.LeaseIdempotencyKey, ttl),
	}, nil
}

// CompleteAsyncCall completes an async call.
//
// Either [response] or [responseError] must be provided, but not both.
func (d *DAL) CompleteAsyncCall(ctx context.Context, call *AsyncCall, response []byte, responseError optional.Option[string]) error {
	if (response == nil) != responseError.Ok() {
		return fmt.Errorf("must provide exactly one of response or error")
	}
	_, err := d.db.CompleteAsyncCall(ctx, response, responseError, call.ID)
	if err != nil {
		return translatePGError(err)
	}
	return nil
}

func (d *DAL) LoadAsyncCall(ctx context.Context, id int64) (*AsyncCall, error) {
	row, err := d.db.LoadAsyncCall(ctx, id)
	if err != nil {
		return nil, translatePGError(err)
	}
	return &AsyncCall{
		ID:        row.ID,
		Verb:      row.Verb,
		Origin:    AsyncCallOrigin(row.Origin),
		OriginKey: row.OriginKey,
		Request:   row.Request,
	}, nil
}
