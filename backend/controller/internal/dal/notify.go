package dal

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/jackc/pgx/v5"
	"github.com/jpillora/backoff"

	log "github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/schema"
)

// NotificationPayload is a row from the database.
//
//sumtype:decl
type NotificationPayload interface{ notification() }

// A Notification from the database.
type Notification[T NotificationPayload, Key any, KeyP interface {
	*Key
	encoding.TextUnmarshaler
}] struct {
	Deleted types.Option[Key] // If present the object was deleted.
	Message T
}

type DeploymentNotification = Notification[Deployment, model.DeploymentName, *model.DeploymentName]

// See JSON structure in SQL schema
type event struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	New    string `json:"new,omitempty"`
	Old    string `json:"old,omitempty"`
}

func (d *DAL) runListener(ctx context.Context, conn *pgx.Conn) {
	logger := log.FromContext(ctx)

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

func (d *DAL) publishNotification(ctx context.Context, notification event, logger *log.Logger) error {
	switch notification.Table {
	case "deployments":
		deployment, err := decodeNotification(notification, func(key model.DeploymentName) (Deployment, error) {
			row, err := d.db.GetDeployment(ctx, key)
			if err != nil {
				return Deployment{}, errors.WithStack(err)
			}
			moduleSchema, err := schema.ModuleFromBytes(row.Schema)
			if err != nil {
				return Deployment{}, errors.WithStack(err)
			}
			return Deployment{
				CreatedAt:   row.CreatedAt.Time,
				Name:        row.Name,
				Module:      row.ModuleName,
				Schema:      moduleSchema,
				MinReplicas: int(row.MinReplicas),
				Language:    row.Language,
			}, nil
		})
		if err != nil {
			return errors.WithStack(err)
		}
		logger.Infof("Deployment notification: %v", deployment.Message.Name)
		d.DeploymentChanges.Publish(deployment)

	default:
		panic(fmt.Sprintf("unknown table %q in DB notification", notification.Table))
	}
	return nil
}

func decodeNotification[Key any, T NotificationPayload, KeyP interface {
	*Key
	encoding.TextUnmarshaler
}](notification event, translate func(key Key) (T, error)) (Notification[T, Key, KeyP], error) {
	var (
		deleted types.Option[Key]
		message T
		err     error
	)
	if notification.Action == "DELETE" {
		var deletedKey Key
		var deletedKeyP KeyP = &deletedKey
		err = deletedKeyP.UnmarshalText([]byte(notification.Old))
		deleted = types.Some(deletedKey)
	} else {
		var newKey Key
		var newKeyP KeyP = &newKey
		err = newKeyP.UnmarshalText([]byte(notification.New))
		if err == nil {
			message, err = translate(newKey)
			if err != nil {
				return Notification[T, Key, KeyP]{}, errors.Wrap(err, "failed to translate database event")
			}
		}
	}
	if err != nil {
		return Notification[T, Key, KeyP]{}, errors.WithStack(err)
	}

	return Notification[T, Key, KeyP]{Deleted: deleted, Message: message}, nil
}

func waitForNotification(ctx context.Context, conn *pgx.Conn) (event, error) {
	notification, err := conn.WaitForNotification(ctx)
	if err != nil {
		return event{}, errors.WithStack(err)
	}
	ev := event{}
	dec := json.NewDecoder(strings.NewReader(notification.Payload))
	dec.DisallowUnknownFields()
	err = dec.Decode(&ev)
	if err != nil {
		return event{}, errors.WithStack(err)
	}
	return ev, nil
}
