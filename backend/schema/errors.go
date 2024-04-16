package schema

import (
	"errors"
	"fmt"
	"sort"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Error struct {
	Msg       string   `json:"msg" protobuf:"1"`
	Pos       Position `json:"pos" protobuf:"2"`
	EndColumn int      `json:"endCol" protobuf:"3"`
}

func (e Error) ToProto() *schemapb.Error {
	return &schemapb.Error{
		Msg:       e.Msg,
		Pos:       posToProto(e.Pos),
		EndColumn: int64(e.EndColumn),
	}
}

func (e Error) Error() string { return fmt.Sprintf("%s-%d: %s", e.Pos, e.EndColumn, e.Msg) }

func errorFromProto(e *schemapb.Error) *Error {
	return &Error{
		Pos:       posFromProto(e.Pos),
		Msg:       e.Msg,
		EndColumn: int(e.EndColumn),
	}
}

func errorsToProto(errs []*Error) []*schemapb.Error {
	var out []*schemapb.Error
	for _, s := range errs {
		pb := s.ToProto()
		out = append(out, pb)
	}
	return out
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

func (e *ErrorList) ToProto() *schemapb.ErrorList {
	return &schemapb.ErrorList{
		Errors: errorsToProto(e.Errors),
	}
}

// ErrorListFromProto converts a protobuf ErrorList to an ErrorList.
func ErrorListFromProto(e *schemapb.ErrorList) *ErrorList {
	return &ErrorList{
		Errors: errorsFromProto(e.Errors),
	}
}

func Errorf(pos Position, endColumn int, format string, args ...any) *Error {
	err := Error{Msg: fmt.Sprintf(format, args...), Pos: pos, EndColumn: endColumn}
	return &err
}

func Wrapf(pos Position, endColumn int, err error, format string, args ...any) *Error {
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
	e := Error{Msg: fmt.Sprintf(format, args...), Pos: newPos, EndColumn: newEndColumn}
	return &e
}

func SortErrorsByPosition(merr []error) {
	sort.Slice(merr, func(i, j int) bool {
		var ipe, jpe Error
		if errors.As(merr[i], &ipe) && errors.As(merr[j], &jpe) {
			ipp := ipe.Pos
			jpp := jpe.Pos
			return ipp.Line < jpp.Line || (ipp.Line == jpp.Line && ipp.Column < jpp.Column) || (ipp.Line == jpp.Line && ipp.Column == jpp.Column && ipe.EndColumn < jpe.EndColumn)
		}
		return merr[i].Error() < merr[j].Error()
	})
}
