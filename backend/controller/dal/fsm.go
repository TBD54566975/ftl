package dal

import (
	"context"
	"encoding/json"

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
