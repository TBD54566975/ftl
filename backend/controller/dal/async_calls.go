package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
		Verb:    verb.String(),
		Request: request,
	})
	return translatePGError(err)
}

// AcquireAsyncCall acquires a pending async call to execute.
//
// Returns ErrNotFound if there are no async calls to acquire.
func (d *DAL) AcquireAsyncCall(ctx context.Context) (*Lease, error) {
	ttl := time.Second * 5
	row, err := d.db.AcquireAsyncCall(ctx, ttl)
	if err != nil {
		err = translatePGError(err)
		// We get a NULL constraint violation if there are no async calls to acquire, so translate it to ErrNotFound.
		if errors.Is(err, ErrConstraint) {
			return nil, fmt.Errorf("no pending async calls: %w", ErrNotFound)
		}
		return nil, err
	}
	return d.newLease(ctx, row.LeaseKey, row.LeaseIdempotencyKey, ttl), nil
}
