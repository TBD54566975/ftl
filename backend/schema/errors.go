package schema

import (
	"errors"
	"fmt"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Error struct {
	Msg       string   `json:"msg" protobuf:"1"`
	Pos       Position `json:"pos" protobuf:"2"`
	EndColumn int      `json:"endCol" protobuf:"3"`
	Err       error    `protobuf:"-"` // Wrapped error, if any
}

func errorFromProto(e *schemapb.Error) *Error {
	return &Error{
		Pos:       posFromProto(e.Pos),
		Msg:       e.Msg,
		EndColumn: int(e.EndColumn),
	}
}

func errorsFromProto(errs []*schemapb.Error) []*Error {
	var out []*Error
	for _, pb := range errs {
		s := errorFromProto(pb)
		out = append(out, s)
	}
	return out
}

type ErrorList struct {
	Errors []*Error `json:"errors" protobuf:"1"`
}

// ErrorListFromProto converts a protobuf ErrorList to an ErrorList.
func ErrorListFromProto(e *schemapb.ErrorList) *ErrorList {
	return &ErrorList{
		Errors: errorsFromProto(e.Errors),
	}
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
