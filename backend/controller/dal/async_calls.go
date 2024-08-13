package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/backend/schema"
)

type asyncOriginParseRoot struct {
	Key AsyncOrigin `parser:"@@"`
}

var asyncOriginParser = participle.MustBuild[asyncOriginParseRoot](
	participle.Union[AsyncOrigin](AsyncOriginFSM{}, AsyncOriginPubSub{}),
)

// AsyncOrigin is a sum type representing the originator of an async call.
//
// This is used to determine how to handle the result of the async call.
type AsyncOrigin interface {
	asyncOrigin()
	// Origin returns the origin type.
	Origin() string
	String() string
}

// AsyncOriginFSM represents the context for the originator of an FSM async call.
//
// It is in the form fsm:<module>.<name>:<key>
type AsyncOriginFSM struct {
	FSM schema.RefKey `parser:"'fsm' ':' @@"`
	Key string        `parser:"':' @(~EOF)+"`
}

var _ AsyncOrigin = AsyncOriginFSM{}

func (AsyncOriginFSM) asyncOrigin()     {}
func (a AsyncOriginFSM) Origin() string { return "fsm" }
func (a AsyncOriginFSM) String() string { return fmt.Sprintf("fsm:%s:%s", a.FSM, a.Key) }

// AsyncOriginPubSub represents the context for the originator of an PubSub async call.
//
// It is in the form fsm:<module>.<subscription_name>
type AsyncOriginPubSub struct {
	Subscription schema.RefKey `parser:"'sub' ':' @@"`
}

var _ AsyncOrigin = AsyncOriginPubSub{}

func (AsyncOriginPubSub) asyncOrigin()     {}
func (a AsyncOriginPubSub) Origin() string { return "sub" }
func (a AsyncOriginPubSub) String() string { return fmt.Sprintf("sub:%s", a.Subscription) }

// ParseAsyncOrigin parses an async origin key.
func ParseAsyncOrigin(origin string) (AsyncOrigin, error) {
	root, err := asyncOriginParser.ParseString("", origin)
	if err != nil {
		return nil, err
	}
	return root.Key, nil
}

type AsyncCall struct {
	*Lease           // May be nil
	ID               int64
	Origin           AsyncOrigin
	Verb             schema.RefKey
	CatchVerb        optional.Option[schema.RefKey]
	Request          json.RawMessage
	ScheduledAt      time.Time
	QueueDepth       int64
	ParentRequestKey optional.Option[string]
	TraceContext     []byte

	Error optional.Option[string]

	RemainingAttempts int32
	Backoff           time.Duration
	MaxBackoff        time.Duration
	Catching          bool
}

// AcquireAsyncCall acquires a pending async call to execute.
//
// Returns ErrNotFound if there are no async calls to acquire.
func (d *DAL) AcquireAsyncCall(ctx context.Context) (call *AsyncCall, err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	ttl := time.Second * 5
	row, err := tx.db.AcquireAsyncCall(ctx, sqltypes.Duration(ttl))
	if err != nil {
		err = dalerrs.TranslatePGError(err)
		if errors.Is(err, dalerrs.ErrNotFound) {
			return nil, fmt.Errorf("no pending async calls: %w", dalerrs.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to acquire async call: %w", err)
	}
	origin, err := ParseAsyncOrigin(row.Origin)
	if err != nil {
		return nil, fmt.Errorf("failed to parse origin key %q: %w", row.Origin, err)
	}

	decryptedRequest, err := d.decrypt(encryption.AsyncSubKey, row.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt async call request: %w", err)
	}

	lease, _ := d.newLease(ctx, row.LeaseKey, row.LeaseIdempotencyKey, ttl)
	return &AsyncCall{
		ID:                row.AsyncCallID,
		Verb:              row.Verb,
		Origin:            origin,
		CatchVerb:         row.CatchVerb,
		Request:           decryptedRequest,
		Lease:             lease,
		ScheduledAt:       row.ScheduledAt,
		QueueDepth:        row.QueueDepth,
		ParentRequestKey:  row.ParentRequestKey,
		TraceContext:      row.TraceContext.RawMessage,
		RemainingAttempts: row.RemainingAttempts,
		Error:             row.Error,
		Backoff:           time.Duration(row.Backoff),
		MaxBackoff:        time.Duration(row.MaxBackoff),
		Catching:          row.Catching,
	}, nil
}

// CompleteAsyncCall completes an async call.
//
// "result" is either a []byte representing the successful response, or a string
// representing a failure message.
func (d *DAL) CompleteAsyncCall(ctx context.Context,
	call *AsyncCall,
	result either.Either[[]byte, string],
	finalise func(tx *Tx, isFinalResult bool) error) (didScheduleAnotherCall bool, err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return false, dalerrs.TranslatePGError(err) //nolint:wrapcheck
	}
	defer tx.CommitOrRollback(ctx, &err)

	isFinalResult := true
	didScheduleAnotherCall = false
	switch result := result.(type) {
	case either.Left[[]byte, string]: // Successful response.
		_, err = tx.db.SucceedAsyncCall(ctx, result.Get(), call.ID)
		if err != nil {
			return false, dalerrs.TranslatePGError(err) //nolint:wrapcheck
		}

	case either.Right[[]byte, string]: // Failure message.
		if call.RemainingAttempts > 0 {
			_, err = d.db.FailAsyncCallWithRetry(ctx, sql.FailAsyncCallWithRetryParams{
				ID:                call.ID,
				Error:             result.Get(),
				RemainingAttempts: call.RemainingAttempts - 1,
				Backoff:           sqltypes.Duration(min(call.Backoff*2, call.MaxBackoff)),
				MaxBackoff:        sqltypes.Duration(call.MaxBackoff),
				ScheduledAt:       time.Now().Add(call.Backoff),
			})
			if err != nil {
				return false, dalerrs.TranslatePGError(err) //nolint:wrapcheck
			}
			isFinalResult = false
			didScheduleAnotherCall = true
		} else if call.RemainingAttempts == 0 && call.CatchVerb.Ok() {
			// original error is the last error that occurred before we started to catch
			originalError := call.Error.Default(result.Get())
			// scheduledAt should be immediate if this is our first catch attempt, otherwise we should use backoff
			scheduledAt := time.Now()
			if call.Catching {
				scheduledAt = scheduledAt.Add(call.Backoff)
			}
			_, err = d.db.FailAsyncCallWithRetry(ctx, sql.FailAsyncCallWithRetryParams{
				ID:                call.ID,
				Error:             result.Get(),
				RemainingAttempts: 0,
				Backoff:           sqltypes.Duration(call.Backoff), // maintain backoff
				MaxBackoff:        sqltypes.Duration(call.MaxBackoff),
				ScheduledAt:       scheduledAt,
				Catching:          true,
				OriginalError:     optional.Some(originalError),
			})
			if err != nil {
				return false, dalerrs.TranslatePGError(err) //nolint:wrapcheck
			}
			isFinalResult = false
			didScheduleAnotherCall = true
		} else {
			_, err = tx.db.FailAsyncCall(ctx, result.Get(), call.ID)
			if err != nil {
				return false, dalerrs.TranslatePGError(err) //nolint:wrapcheck
			}
		}
	}
	if err := finalise(tx, isFinalResult); err != nil {
		return false, err
	}
	return didScheduleAnotherCall, nil
}

func (d *DAL) LoadAsyncCall(ctx context.Context, id int64) (*AsyncCall, error) {
	row, err := d.db.LoadAsyncCall(ctx, id)
	if err != nil {
		return nil, dalerrs.TranslatePGError(err)
	}
	origin, err := ParseAsyncOrigin(row.Origin)
	if err != nil {
		return nil, fmt.Errorf("failed to parse origin key %q: %w", row.Origin, err)
	}
	return &AsyncCall{
		ID:      row.ID,
		Verb:    row.Verb,
		Origin:  origin,
		Request: row.Request,
	}, nil
}
