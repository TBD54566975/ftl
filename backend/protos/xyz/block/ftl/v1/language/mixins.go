package languagepb

import (
	"fmt"

	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/slices"
)

// ErrorsFromProto converts a protobuf ErrorList to a []builderrors.Error.
func ErrorsFromProto(e *ErrorList) []builderrors.Error {
	return slices.Map(e.Errors, errorFromProto)
}

func ErrorsToProto(errs []builderrors.Error) *ErrorList {
	return &ErrorList{Errors: slices.Map(errs, errorToProto)}
}

func levelFromProto(level Error_ErrorLevel) builderrors.ErrorLevel {
	switch level {
	case Error_INFO:
		return builderrors.INFO
	case Error_WARN:
		return builderrors.WARN
	case Error_ERROR:
		return builderrors.ERROR
	}
	panic(fmt.Sprintf("unhandled ErrorLevel %v", level))
}

func levelToProto(level builderrors.ErrorLevel) Error_ErrorLevel {
	switch level {
	case builderrors.INFO:
		return Error_INFO
	case builderrors.WARN:
		return Error_WARN
	case builderrors.ERROR:
		return Error_ERROR
	}
	panic(fmt.Sprintf("unhandled ErrorLevel %v", level))
}

func errorFromProto(e *Error) builderrors.Error {
	return builderrors.Error{
		Pos:   PosFromProto(e.Pos),
		Msg:   e.Msg,
		Level: levelFromProto(e.Level),
	}
}

func errorToProto(e builderrors.Error) *Error {
	return &Error{
		Msg: e.Msg,
		Pos: &Position{
			Filename:    e.Pos.Filename,
			StartColumn: int64(e.Pos.StartColumn),
			EndColumn:   int64(e.Pos.EndColumn),
			Line:        int64(e.Pos.Line),
		},
		Level: levelToProto(e.Level),
	}
}

func PosFromProto(pos *Position) builderrors.Position {
	if pos == nil {
		return builderrors.Position{}
	}
	return builderrors.Position{
		Line:        int(pos.Line),
		StartColumn: int(pos.StartColumn),
		EndColumn:   int(pos.EndColumn),
		Filename:    pos.Filename,
	}
}
