package dal

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/jackc/pgx/v5"
	"github.com/jpillora/backoff"

	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
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
	Deleted optional.Option[Key] // If present the object was deleted.
	Message optional.Option[T]
}

func (n Notification[T, Key, KeyP]) String() string {
	if key, ok := n.Deleted.Get(); ok {
		return fmt.Sprintf("deleted %v", key)
	}
	return fmt.Sprintf("message %v", n.Message)
}

// DeploymentNotification is a notification from the database when a deployment changes.
type DeploymentNotification = Notification[Deployment, model.DeploymentKey, *model.DeploymentKey]

// See JSON structure in SQL schema
type event struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	New    string `json:"new,omitempty"`
	Old    string `json:"old,omitempty"`
	// Optional field for conveying deletion metadata.
	Deleted json.RawMessage `json:"deleted,omitempty"`
}

func (d *DAL) runListener(ctx context.Context, conn *pgx.Conn) {
	defer conn.Close(ctx)
	logger := log.FromContext(ctx)

	logger.Debugf("Starting DB listener")

	// Wait for the notification channel to be ready.
	retry := backoff.Backoff{}
	channels := []string{"deployments_events", "topics_events", "topic_events_events"}
	for _, channel := range channels {
		for {
			_, err := conn.Exec(ctx, "LISTEN "+channel)
			if err == nil {
				logger.Debugf("Listening to channel: %s", channel)
				retry.Reset()
				break
			}
			logger.Errorf(err, "Failed to LISTEN to %s", channel)
			time.Sleep(retry.Duration())
		}
	}

	// Main loop for listening to notifications.
	for {
		delay := time.Millisecond * 100
		logger.Tracef("Waiting for notification")
		notification, err := waitForNotification(ctx, conn)
		if err == nil {
			logger.Tracef("Publishing notification: %s", notification)
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
		deployment, err := decodeNotification(notification, func(key model.DeploymentKey) (Deployment, optional.Option[model.DeploymentKey], error) {
			row, err := d.db.GetDeployment(ctx, key)
			if err != nil {
				return Deployment{}, optional.None[model.DeploymentKey](), dalerrs.TranslatePGError(err)
			}
			return Deployment{
				CreatedAt:   row.Deployment.CreatedAt,
				Key:         row.Deployment.Key,
				Module:      row.ModuleName,
				Schema:      row.Deployment.Schema,
				MinReplicas: int(row.Deployment.MinReplicas),
				Language:    row.Language,
			}, optional.None[model.DeploymentKey](), nil
		})
		if err != nil {
			return err
		}
		logger.Tracef("Deployment notification: %s", deployment)
		d.DeploymentChanges.Publish(deployment)

	case "topics":
		// TODO: handle topics notifications
	case "topic_events":
		// TODO: handle topic events notifications
	default:
		panic(fmt.Sprintf("unknown table %q in DB notification", notification.Table))
	}
	return nil
}

// This function takes a notification from the database and translates it into
// a concrete Notification value.
//
// The translate function is called to translate the key into a concrete value
// OR a delete notification.
func decodeNotification[K any, T NotificationPayload, KP interface {
	*K
	encoding.TextUnmarshaler
}](notification event, translate func(key K) (T, optional.Option[K], error)) (Notification[T, K, KP], error) {
	var (
		deleted optional.Option[K]
		message optional.Option[T]
	)
	if notification.Action == "DELETE" {
		var deletedKey K
		var deletedKeyP KP = &deletedKey
		if err := deletedKeyP.UnmarshalText([]byte(notification.Old)); err != nil {
			return Notification[T, K, KP]{}, fmt.Errorf("failed to unmarshal notification key: %w", err)
		}
		deleted = optional.Some(deletedKey)
	} else {
		var newKey K
		var newKeyP KP = &newKey
		if err := newKeyP.UnmarshalText([]byte(notification.New)); err != nil {
			return Notification[T, K, KP]{}, fmt.Errorf("failed to unmarshal notification key: %w", err)
		}
		var msg T
		var err error
		msg, deleted, err = translate(newKey)
		if err != nil {
			return Notification[T, K, KP]{}, fmt.Errorf("failed to translate database notification: %w", err)
		}

		if !deleted.Ok() {
			message = optional.Some(msg)
		}
	}

	return Notification[T, K, KP]{Deleted: deleted, Message: message}, nil
}

func waitForNotification(ctx context.Context, conn *pgx.Conn) (event, error) {
	notification, err := conn.WaitForNotification(ctx)
	if err != nil {
		return event{}, err
	}
	ev := event{}
	dec := json.NewDecoder(strings.NewReader(notification.Payload))
	dec.DisallowUnknownFields()
	err = dec.Decode(&ev)
	if err != nil {
		return event{}, err
	}
	return ev, nil
}
