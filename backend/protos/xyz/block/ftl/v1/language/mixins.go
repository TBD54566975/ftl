package languagepb

import (
	"fmt"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
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
		Pos:   schema.PosFromProto(e.Pos).ToErrorPosWithEnd(int(e.EndColumn)),
		Msg:   e.Msg,
		Level: levelFromProto(e.Level),
	}
}

func errorToProto(e builderrors.Error) *Error {
	return &Error{
		Msg: e.Msg,
		Pos: &schemapb.Position{
			Filename: e.Pos.Filename,
			Column:   int64(e.Pos.StartColumn),
			Line:     int64(e.Pos.Line),
		},
		EndColumn: int64(e.Pos.EndColumn),
		Level:     levelToProto(e.Level),
	}
}
