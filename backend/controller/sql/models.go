// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package sql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
)

type ControllerState string

const (
	ControllerStateLive ControllerState = "live"
	ControllerStateDead ControllerState = "dead"
)

func (e *ControllerState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ControllerState(s)
	case string:
		*e = ControllerState(s)
	default:
		return fmt.Errorf("unsupported scan type for ControllerState: %T", src)
	}
	return nil
}

type NullControllerState struct {
	ControllerState ControllerState
	Valid           bool // Valid is true if ControllerState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullControllerState) Scan(value interface{}) error {
	if value == nil {
		ns.ControllerState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ControllerState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullControllerState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ControllerState), nil
}

type EventType string

const (
	EventTypeCall              EventType = "call"
	EventTypeLog               EventType = "log"
	EventTypeDeploymentCreated EventType = "deployment_created"
	EventTypeDeploymentUpdated EventType = "deployment_updated"
)

func (e *EventType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EventType(s)
	case string:
		*e = EventType(s)
	default:
		return fmt.Errorf("unsupported scan type for EventType: %T", src)
	}
	return nil
}

type NullEventType struct {
	EventType EventType
	Valid     bool // Valid is true if EventType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEventType) Scan(value interface{}) error {
	if value == nil {
		ns.EventType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EventType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEventType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EventType), nil
}

type Origin string

const (
	OriginIngress Origin = "ingress"
	OriginCron    Origin = "cron"
	OriginPubsub  Origin = "pubsub"
)

func (e *Origin) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Origin(s)
	case string:
		*e = Origin(s)
	default:
		return fmt.Errorf("unsupported scan type for Origin: %T", src)
	}
	return nil
}

type NullOrigin struct {
	Origin Origin
	Valid  bool // Valid is true if Origin is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullOrigin) Scan(value interface{}) error {
	if value == nil {
		ns.Origin, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Origin.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullOrigin) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Origin), nil
}

type RunnerState string

const (
	RunnerStateIdle     RunnerState = "idle"
	RunnerStateReserved RunnerState = "reserved"
	RunnerStateAssigned RunnerState = "assigned"
	RunnerStateDead     RunnerState = "dead"
)

func (e *RunnerState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RunnerState(s)
	case string:
		*e = RunnerState(s)
	default:
		return fmt.Errorf("unsupported scan type for RunnerState: %T", src)
	}
	return nil
}

type NullRunnerState struct {
	RunnerState RunnerState
	Valid       bool // Valid is true if RunnerState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRunnerState) Scan(value interface{}) error {
	if value == nil {
		ns.RunnerState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RunnerState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRunnerState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.RunnerState), nil
}

type Artefact struct {
	ID        int64
	CreatedAt time.Time
	Digest    []byte
	Content   []byte
}

type Controller struct {
	ID       int64
	Key      model.ControllerKey
	Created  time.Time
	LastSeen time.Time
	State    ControllerState
	Endpoint string
}

type Deployment struct {
	ID          int64
	CreatedAt   time.Time
	ModuleID    int64
	Key         model.DeploymentKey
	Schema      *schema.Module
	Labels      []byte
	MinReplicas int32
}

type DeploymentArtefact struct {
	ArtefactID   int64
	DeploymentID int64
	CreatedAt    time.Time
	Executable   bool
	Path         string
}

type Event struct {
	ID           int64
	TimeStamp    time.Time
	DeploymentID int64
	RequestID    optional.Option[int64]
	Type         EventType
	CustomKey1   optional.Option[string]
	CustomKey2   optional.Option[string]
	CustomKey3   optional.Option[string]
	CustomKey4   optional.Option[string]
	Payload      json.RawMessage
}

type IngressRoute struct {
	Method       string
	Path         string
	DeploymentID int64
	Module       string
	Verb         string
}

type Module struct {
	ID       int64
	Language string
	Name     string
}

type Request struct {
	ID         int64
	Origin     Origin
	Name       string
	SourceAddr string
}

type Runner struct {
	ID                 int64
	Key                model.RunnerKey
	Created            time.Time
	LastSeen           time.Time
	ReservationTimeout NullTime
	State              RunnerState
	Endpoint           string
	ModuleName         optional.Option[string]
	DeploymentID       optional.Option[int64]
	Labels             []byte
}

type Topic struct {
	ID        int64
	Key       interface{}
	CreatedAt time.Time
	ModuleID  int64
	Name      string
	Type      string
}

type TopicEvent struct {
	ID        int64
	CreatedAt time.Time
	TopicID   int64
	Payload   []byte
}

type TopicSubscriber struct {
	ID                   int64
	Key                  interface{}
	CreatedAt            time.Time
	TopicSubscriptionsID int64
	DeploymentID         int64
	Verb                 string
}

type TopicSubscription struct {
	ID        int64
	Key       interface{}
	CreatedAt time.Time
	TopicID   int64
	Name      string
	Cursor    int64
}
