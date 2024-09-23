package builderrors

import (
	"errors"
	"fmt"
	"sort"
)

type Position struct {
	Filename    string
	Offset      int
	StartColumn int
	EndColumn   int
	Line        int
}

func (p Position) String() string {
	columnStr := fmt.Sprintf("%d", p.StartColumn)
	if p.StartColumn != p.EndColumn {
		columnStr += fmt.Sprintf("-%d", p.EndColumn)
	}
	if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, columnStr)
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, columnStr)
}

type ErrorLevel int

const (
	INFO ErrorLevel = iota
	WARN
	ERROR
)

type Error struct {
	Msg   string
	Pos   Position
	Level ErrorLevel
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }

func makeError(level ErrorLevel, pos Position, format string, args ...any) *Error {
	err := Error{Msg: fmt.Sprintf(format, args...), Pos: pos, Level: level}
	return &err
}

func Infof(pos Position, format string, args ...any) *Error {
	return makeError(INFO, pos, format, args...)
}

func Warnf(pos Position, format string, args ...any) *Error {
	return makeError(WARN, pos, format, args...)
}

func Errorf(pos Position, format string, args ...any) *Error {
	return makeError(ERROR, pos, format, args...)
}

func Wrapf(pos Position, endColumn int, err error, format string, args ...any) *Error {
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
	return makeError(ERROR, newPos, format, args...)
}

func SortErrorsByPosition(merr []*Error) {
	if merr == nil {
		return
	}
	sort.Slice(merr, func(i, j int) bool {
		ipp := merr[i].Pos
		jpp := merr[j].Pos
		return ipp.Line < jpp.Line || (ipp.Line == jpp.Line && ipp.StartColumn < jpp.StartColumn) ||
			(ipp.Line == jpp.Line && ipp.StartColumn == jpp.StartColumn && merr[i].Pos.EndColumn < merr[j].Pos.EndColumn) ||
			(ipp.Line == jpp.Line && ipp.StartColumn == jpp.StartColumn && merr[i].Pos.EndColumn == merr[j].Pos.EndColumn && merr[i].Msg < merr[j].Msg)
	})
}

func ContainsTerminalError(errs []*Error) bool {
	for _, e := range errs {
		if e.Level == ERROR {
			return true
		}
	}
	return false
}
