package schema

import (
	"errors"
	"fmt"
)

type Error struct {
	Msg       string   `json:"msg" protobuf:"1"`
	Pos       Position `json:"pos" protobuf:"2"`
	EndColumn int      `json:"endCol" protobuf:"3"`
	Err       error    `protobuf:"-"` // Wrapped error, if any
}

type ErrorList struct {
	Errors []Error `json:"errors" protobuf:"1"`
}

func (e Error) Error() string { return fmt.Sprintf("%s-%d: %s", e.Pos, e.EndColumn, e.Msg) }
func (e Error) Unwrap() error { return e.Err }

func Errorf(pos Position, endColumn int, format string, args ...any) Error {
	return Error{Msg: fmt.Sprintf(format, args...), Pos: pos, EndColumn: endColumn}
}

func Wrapf(pos Position, endColumn int, err error, format string, args ...any) Error {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	// Propagate existing error position if available
	var newPos Position
	var newEndColumn int
	if perr := (Error{}); errors.As(err, &perr) {
		newPos = perr.Pos
		newEndColumn = perr.EndColumn
		args = append(args, perr.Msg)
	} else {
		newPos = pos
		newEndColumn = endColumn
		args = append(args, err)
	}
	return Error{Msg: fmt.Sprintf(format, args...), Pos: newPos, EndColumn: newEndColumn, Err: err}
}
