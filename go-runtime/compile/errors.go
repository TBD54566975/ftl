package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/TBD54566975/ftl/backend/schema"
)

type Error struct {
	Msg string
	Pos schema.Position
	Err error // Wrapped error, if any
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }
func (e Error) Unwrap() error { return e.Err }

func errorf(pos token.Pos, format string, args ...any) Error {
	return Error{Msg: fmt.Sprintf(format, args...), Pos: goPosToSchemaPos(pos)}
}

func wrapf(node ast.Node, err error, format string, args ...any) Error {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	// Propagate existing error position if available
	var pos schema.Position
	if perr := (Error{}); errors.As(err, &perr) {
		pos = perr.Pos
		args = append(args, perr.Msg)
	} else {
		pos = goPosToSchemaPos(node.Pos())
		args = append(args, err)
	}
	return Error{Msg: fmt.Sprintf(format, args...), Pos: pos, Err: err}
}
