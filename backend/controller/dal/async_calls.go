package dal

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dal/internal/sql"
	leasedal "github.com/TBD54566975/ftl/backend/controller/leases/dal"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/model"
)

type asyncOriginParseRoot struct {
	Key AsyncOrigin `parser:"@@"`
}

var asyncOriginLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"NumberIdent", `[0-9][a-zA-Z0-9_-]*`},
	{"Ident", `[a-zA-Z_][a-zA-Z0-9_-]*`},
	{"Punct", `[:.]`},
})

var asyncOriginParser = participle.MustBuild[asyncOriginParseRoot](
	participle.Union[AsyncOrigin](AsyncOriginCron{}, AsyncOriginFSM{}, AsyncOriginPubSub{}),
	participle.Lexer(asyncOriginLexer),
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

// AsyncOriginCron represents the context for the originator of a cron async call.
//
// It is in the form cron:<module>.<verb>
type AsyncOriginCron struct {
	CronJobKey model.CronJobKey `parser:"'cron' ':' @(~EOF)+"`
}

var _ AsyncOrigin = AsyncOriginCron{}

func (AsyncOriginCron) asyncOrigin()     {}
func (a AsyncOriginCron) Origin() string { return "cron" }
func (a AsyncOriginCron) String() string { return fmt.Sprintf("cron:%s", a.CronJobKey) }

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
	*leasedal.Lease  // May be nil
	ID               int64
	Origin           AsyncOrigin
	Verb             schema.RefKey
	CatchVerb        optional.Option[schema.RefKey]
	Request          []byte
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
func (d *DAL) AcquireAsyncCall(ctx context.Context) (call *AsyncCall, leaseCtx context.Context, err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return nil, ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	ttl := time.Second * 5
	row, err := tx.db.AcquireAsyncCall(ctx, sqltypes.Duration(ttl))
	if err != nil {
		err = libdal.TranslatePGError(err)
		if errors.Is(err, libdal.ErrNotFound) {
			return nil, ctx, fmt.Errorf("no pending async calls: %w", libdal.ErrNotFound)
		}
		return nil, ctx, fmt.Errorf("failed to acquire async call: %w", err)
	}
	origin, err := ParseAsyncOrigin(row.Origin)
	if err != nil {
		return nil, ctx, fmt.Errorf("failed to parse origin key %q: %w", row.Origin, err)
	}

	fmt.Println("row.Request", row.Request)

	decryptedRequest, err := d.encryption.Decrypt(&row.Request)
	if err != nil {
		return nil, ctx, fmt.Errorf("failed to decrypt async call request: %w", err)
	}

	lease, leaseCtx := d.leaseDAL.NewLease(ctx, row.LeaseKey, row.LeaseIdempotencyKey, ttl)
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
	}, leaseCtx, nil
}

// CompleteAsyncCall completes an async call.
// The call will use the existing transaction if d is a transaction. Otherwise it will create and commit a new transaction.
//
// "result" is either a []byte representing the successful response, or a string
// representing a failure message.
func (d *DAL) CompleteAsyncCall(ctx context.Context,
	call *AsyncCall,
	result either.Either[[]byte, string],
	finalise func(tx *DAL, isFinalResult bool) error) (didScheduleAnotherCall bool, err error) {
	var tx *DAL
	switch d.Connection.(type) {
	case *dbsql.DB:
		tx, err = d.Begin(ctx)
		if err != nil {
			return false, libdal.TranslatePGError(err) //nolint:wrapcheck
		}
		defer tx.CommitOrRollback(ctx, &err)
	case *dbsql.Tx:
		tx = d
	default:
		return false, errors.New("invalid connection type")
	}

	isFinalResult := true
	didScheduleAnotherCall = false
	switch result := result.(type) {
	case either.Left[[]byte, string]: // Successful response.
		var encryptedResult encryption.EncryptedAsyncColumn
		err := tx.encryption.Encrypt(result.Get(), &encryptedResult)
		if err != nil {
			return false, fmt.Errorf("failed to encrypt async call result: %w", err)
		}
		_, err = tx.db.SucceedAsyncCall(ctx, optional.Some(encryptedResult), call.ID)
		if err != nil {
			return false, libdal.TranslatePGError(err) //nolint:wrapcheck
		}

	case either.Right[[]byte, string]: // Failure message.
		if call.RemainingAttempts > 0 {
			_, err = tx.db.FailAsyncCallWithRetry(ctx, sql.FailAsyncCallWithRetryParams{
				ID:                call.ID,
				Error:             result.Get(),
				RemainingAttempts: call.RemainingAttempts - 1,
				Backoff:           sqltypes.Duration(min(call.Backoff*2, call.MaxBackoff)),
				MaxBackoff:        sqltypes.Duration(call.MaxBackoff),
				ScheduledAt:       time.Now().Add(call.Backoff),
			})
			if err != nil {
				return false, libdal.TranslatePGError(err) //nolint:wrapcheck
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
			_, err = tx.db.FailAsyncCallWithRetry(ctx, sql.FailAsyncCallWithRetryParams{
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
				return false, libdal.TranslatePGError(err) //nolint:wrapcheck
			}
			isFinalResult = false
			didScheduleAnotherCall = true
		} else {
			_, err = tx.db.FailAsyncCall(ctx, result.Get(), call.ID)
			if err != nil {
				return false, libdal.TranslatePGError(err) //nolint:wrapcheck
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
		return nil, libdal.TranslatePGError(err)
	}
	origin, err := ParseAsyncOrigin(row.Origin)
	if err != nil {
		return nil, fmt.Errorf("failed to parse origin key %q: %w", row.Origin, err)
	}
	request, err := d.encryption.Decrypt(&row.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt async call request: %w", err)
	}
	return &AsyncCall{
		ID:      row.ID,
		Verb:    row.Verb,
		Origin:  origin,
		Request: request,
	}, nil
}

func (d *DAL) GetZombieAsyncCalls(ctx context.Context, limit int) ([]*AsyncCall, error) {
	rows, err := d.db.GetZombieAsyncCalls(ctx, int32(limit))
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	var calls []*AsyncCall
	for _, row := range rows {
		origin, err := ParseAsyncOrigin(row.Origin)
		if err != nil {
			return nil, fmt.Errorf("failed to parse origin key %q: %w", row.Origin, err)
		}
		decryptedRequest, err := d.encryption.Decrypt(&row.Request)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt async call request: %w", err)
		}
		calls = append(calls, &AsyncCall{
			ID:                row.ID,
			Origin:            origin,
			ScheduledAt:       row.ScheduledAt,
			Verb:              row.Verb,
			CatchVerb:         row.CatchVerb,
			Request:           decryptedRequest,
			ParentRequestKey:  row.ParentRequestKey,
			TraceContext:      row.TraceContext.RawMessage,
			Error:             row.Error,
			RemainingAttempts: row.RemainingAttempts,
			Backoff:           time.Duration(row.Backoff),
			MaxBackoff:        time.Duration(row.MaxBackoff),
			Catching:          row.Catching,
		})
	}
	return calls, nil
}
