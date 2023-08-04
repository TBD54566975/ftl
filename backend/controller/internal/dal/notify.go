package dal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/jackc/pgx/v5"
	"github.com/jpillora/backoff"

	log2 "github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/controller/internal/sqltypes"
	"github.com/TBD54566975/ftl/backend/schema"
)

// NotificationPayload is a row from the database.
//
//sumtype:decl
type NotificationPayload interface{ notification() }

// A Notification from the database.
type Notification[T NotificationPayload, Key keyConstraint] struct {
	Deleted types.Option[Key] // If present the object was deleted.
	Message T
}

type DeploymentNotification = Notification[Deployment, model.DeploymentKey]

// See JSON structure in SQL schema
type event struct {
	Table  string       `json:"table"`
	Action string       `json:"action"`
	New    sqltypes.Key `json:"new,omitempty"`
	Old    sqltypes.Key `json:"old,omitempty"`
}

func (d *DAL) runListener(ctx context.Context, conn *pgx.Conn) {
	logger := log2.FromContext(ctx)

	logger.Infof("Starting DB listener")

	// Wait for the notification channel to be ready.
	retry := backoff.Backoff{}
	for {
		_, err := conn.Exec(ctx, "LISTEN notify_events")
		if err == nil {
			break
		}
		logger.Errorf(err, "failed to LISTEN notify_events")
		time.Sleep(retry.Duration())
	}
	retry.Reset()

	// Main loop for listening to notifications.
	for {
		delay := time.Millisecond * 100
		notification, err := waitForNotification(ctx, conn)
		if err == nil {
			err = d.publishNotification(ctx, notification, logger)
		}
		if err != nil {
			logger.Errorf(err, "Failed to receive notification")
			delay = retry.Duration()
		} else {
			retry.Reset()
		}
		select {
		case <-ctx.Done():
			return

		case <-time.After(delay):
		}
	}
}

func (d *DAL) publishNotification(ctx context.Context, notification event, logger *log2.Logger) error {
	switch notification.Table {
	case "deployments":
		deployment, err := decodeNotification(notification, func(key model.DeploymentKey) (Deployment, error) {
			row, err := d.db.GetDeployment(ctx, sqltypes.Key(key))
			if err != nil {
				return Deployment{}, errors.WithStack(err)
			}
			moduleSchema, err := schema.ModuleFromBytes(row.Schema)
			if err != nil {
				return Deployment{}, errors.WithStack(err)
			}
			return Deployment{
				CreatedAt:   row.CreatedAt.Time,
				Key:         model.DeploymentKey(row.Key),
				Module:      row.ModuleName,
				Schema:      moduleSchema,
				MinReplicas: int(row.MinReplicas),
				Language:    row.Language,
			}, nil
		})
		if err != nil {
			return errors.WithStack(err)
		}
		logger.Infof("Deployment notification: %v", deployment.Message.Key)
		d.DeploymentChanges.Publish(deployment)

	default:
		panic(fmt.Sprintf("unknown table %q in DB notification", notification.Table))
	}
	return nil
}

type keyConstraint interface{ ~[16]byte }

func decodeNotification[Key keyConstraint, T NotificationPayload](notification event, translate func(key Key) (T, error)) (Notification[T, Key], error) {
	var (
		deleted types.Option[Key]
		message T
		err     error
	)
	if notification.Action == "DELETE" {
		deleted = types.Some(Key(notification.Old))
	} else {
		message, err = translate(Key(notification.New))
		if err != nil {
			return Notification[T, Key]{}, errors.Wrap(err, "failed to translate database event")
		}
	}

	return Notification[T, Key]{Deleted: deleted, Message: message}, nil
}

func waitForNotification(ctx context.Context, conn *pgx.Conn) (event, error) {
	notification, err := conn.WaitForNotification(ctx)
	if err != nil {
		return event{}, errors.WithStack(err)
	}
	ev := event{}
	err = json.Unmarshal([]byte(notification.Payload), &ev)
	if err != nil {
		return event{}, errors.WithStack(err)
	}
	return ev, nil
}
