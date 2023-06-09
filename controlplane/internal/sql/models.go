// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package sql

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/TBD54566975/ftl/controlplane/internal/sqltypes"
)

type RunnerState string

const (
	RunnerStateIdle     RunnerState = "idle"
	RunnerStateClaimed  RunnerState = "claimed"
	RunnerStateReserved RunnerState = "reserved"
	RunnerStateAssigned RunnerState = "assigned"
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
	CreatedAt pgtype.Timestamptz
	Digest    []byte
	Content   []byte
}

type Deployment struct {
	ID        int64
	CreatedAt pgtype.Timestamptz
	ModuleID  int64
	Key       sqltypes.Key
	Schema    []byte
}

type DeploymentArtefact struct {
	ArtefactID   int64
	DeploymentID int64
	CreatedAt    pgtype.Timestamptz
	Executable   bool
	Path         string
}

type DeploymentLog struct {
	ID           int64
	DeploymentID int64
	Verb         pgtype.Text
	TimeStamp    pgtype.Timestamptz
	Level        int32
	Scope        string
	Message      string
	Error        pgtype.Text
}

type Module struct {
	ID       int64
	Language string
	Name     string
}

type Runner struct {
	ID                 int64
	Key                sqltypes.Key
	LastSeen           pgtype.Timestamptz
	ReservationTimeout pgtype.Timestamptz
	State              RunnerState
	Language           string
	Endpoint           string
	DeploymentID       pgtype.Int8
}
