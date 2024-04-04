package schema

import (
	"errors"
	"fmt"
)

type Error struct {
	Msg string   `json:"msg" protobuf:"1"`
	Pos Position `json:"pos" protobuf:"2"`
	Err error    `protobuf:"-"` // Wrapped error, if any
}

type ErrorList struct {
	Errors []Error `json:"errors" protobuf:"1"`
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }
func (e Error) Unwrap() error { return e.Err }

func Errorf(pos Position, format string, args ...any) Error {
	return Error{Msg: fmt.Sprintf(format, args...), Pos: pos}
}

func Wrapf(pos Position, err error, format string, args ...any) Error {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	// Propagate existing error position if available
	var newPos Position
	if perr := (Error{}); errors.As(err, &perr) {
		newPos = perr.Pos
		args = append(args, perr.Msg)
	} else {
		newPos = pos
		args = append(args, err)
	}
	return Error{Msg: fmt.Sprintf(format, args...), Pos: newPos, Err: err}
}
