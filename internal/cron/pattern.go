package cron

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Whitespace", Pattern: `\s+`},
		{Name: "Ident", Pattern: `\b[a-zA-Z_][a-zA-Z0-9_]*\b`},
		{Name: "Comment", Pattern: `//.*`},
		{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
		{Name: "Number", Pattern: `[0-9]+(?:\.[0-9]+)?`},
		{Name: "Punct", Pattern: `[%/\-\_:[\]{}<>()*+?.,\\^$|#~!\'@]`},
	})

	parserOptions = []participle.Option{
		participle.Lexer(lex),
		participle.Elide("Whitespace"),
		participle.Unquote(),
		participle.Map(func(token lexer.Token) (lexer.Token, error) {
			token.Value = strings.TrimSpace(strings.TrimPrefix(token.Value, "//"))
			return token, nil
		}, "Comment"),
	}

	parser = participle.MustBuild[Pattern](parserOptions...)
)

type Pattern struct {
	Components []Component `parser:"@@*"`
}

func (p Pattern) String() string {
	return strings.Join(slices.Map(p.Components, func(component Component) string {
		return component.String()
	}), " ")
}

func (p Pattern) standardizedComponents() ([]Component, error) {
	switch len(p.Components) {
	case 5:
		// Convert "a b c d e" -> "0 a b c d e *"
		components := make([]Component, 7)
		components[0] = newComponentWithValue(0)
		copy(components[1:], p.Components)
		components[6] = newComponentWithFullRange()
		return components, nil
	case 6:
		// Might be two different formats unfortunately.
		// Could be:
		// - seconds, minutes, hours, day of month, month, day of week
		// - minutes, hours, day of month, month, day of week, year
		// Detect by looking for 4 digit numbers in the last component, and then treat it as a year column
		if isComponentLikelyToBeYearComponent(p.Components[5]) {
			// Convert "a b c d e f" -> "0 a b c d e f"
			components := make([]Component, 7)
			components[0] = newComponentWithValue(0)
			copy(components[1:], p.Components)
			return components, nil
		} else {
			// Convert "a b c d e f" -> "a b c d e f *"
			components := make([]Component, 7)
			copy(components[0:], p.Components)
			components[6] = newComponentWithFullRange()
			return components, nil
		}
	case 7:
		return p.Components, nil
	default:
		return nil, fmt.Errorf("expected 5-7 components, got %d", len(p.Components))
	}
}

func isComponentLikelyToBeYearComponent(component Component) bool {
	for _, s := range component.List {
		if s.ValueRange.Start != nil && *s.ValueRange.Start >= 1000 {
			return true
		}
		if s.ValueRange.End != nil && *s.ValueRange.End >= 1000 {
			return true
		}
	}
	return false
}

type Component struct {
	List []Step `parser:"(@@ (',' @@)*)"`
}

func newComponentWithFullRange() Component {
	return Component{
		List: []Step{
			{
				ValueRange: ValueRange{IsFullRange: true},
			},
		},
	}
}

func newComponentWithValue(value int) Component {
	return Component{
		List: []Step{
			newStepWithValue(value),
		},
	}
}

func (c Component) String() string {
	return strings.Join(slices.Map(c.List, func(step Step) string {
		return step.String()
	}), ",")
}

type Step struct {
	ValueRange ValueRange `parser:"@@"`
	Step       *int       `parser:"('/' @Number)?"`
}

func newStepWithValue(value int) Step {
	return Step{
		ValueRange: ValueRange{Start: &value, End: nil},
	}
}

func (s *Step) String() string {
	if s.Step != nil {
		return fmt.Sprintf("%s/%d", s.ValueRange.String(), *s.Step)
	}
	return s.ValueRange.String()
}

type ValueRange struct {
	IsFullRange bool `parser:"(@'*'"`
	Start       *int `parser:"| @Number"`
	End         *int `parser:"('-' @Number)?)"`
}

func (r *ValueRange) String() string {
	if r.IsFullRange {
		return "*"
	}
	if r.End != nil {
		return fmt.Sprintf("%d-%d", *r.Start, *r.End)
	}
	return strconv.Itoa(*r.Start)
}

func Parse(text string) (Pattern, error) {
	pattern, err := parser.ParseString("", text)
	if err != nil {
		return Pattern{}, err
	}
	return *pattern, nil
}
