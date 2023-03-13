package schema

import (
	"fmt"
	"strings"
)

// This file contains the JSON marshalling and unmarshalling logic as well as
// support methods for visiting as well as type safety.

var _ Type = (*Int)(nil)

func (Int) schemaChildren() []Node { return nil }
func (Int) schemaType()            {}
func (Int) String() string         { return "int" }

var _ Type = (*Float)(nil)

func (Float) schemaChildren() []Node { return nil }
func (Float) schemaType()            {}
func (Float) String() string         { return "float" }

var _ Type = (*String)(nil)

func (String) schemaChildren() []Node { return nil }
func (String) schemaType()            {}
func (String) String() string         { return "string" }

var _ Type = (*Bool)(nil)

func (Bool) schemaChildren() []Node { return nil }
func (Bool) schemaType()            {}
func (Bool) String() string         { return "bool" }

var _ Type = (*Array)(nil)

func (a Array) schemaChildren() []Node { return []Node{a.Element} }
func (Array) schemaType()              {}
func (a Array) String() string         { return "array<" + a.Element.String() + ">" }

var _ Type = (*Map)(nil)

func (m Map) schemaChildren() []Node { return []Node{m.Key, m.Value} }
func (Map) schemaType()              {}
func (m Map) String() string         { return fmt.Sprintf("map<%s, %s>", m.Key.String(), m.Value.String()) }

var _ Node = (*Field)(nil)

func (f Field) schemaChildren() []Node { return []Node{f.Type} }
func (f Field) String() string         { return fmt.Sprintf("%s %s", f.Name, f.Type.String()) }

var _ Type = (*DataRef)(nil)

func (DataRef) schemaChildren() []Node { return nil }
func (DataRef) schemaType()            {}
func (s DataRef) String() string       { return s.Name }

var _ Node = (*Data)(nil)

func (d Data) schemaChildren() []Node {
	children := make([]Node, len(d.Fields))
	for i, f := range d.Fields {
		children[i] = f
	}
	return children
}
func (d Data) String() string {
	out := &strings.Builder{}
	fmt.Fprintf(out, "data %s {\n", d.Name)
	for _, f := range d.Fields {
		fmt.Fprintln(out, indent(f.String()))
	}
	fmt.Fprintf(out, "}")
	return out.String()
}

var _ Type = (*VerbRef)(nil)

func (VerbRef) schemaChildren() []Node { return nil }
func (VerbRef) schemaType()            {}
func (v VerbRef) String() string       { return fmt.Sprintf("%s.%s", v.Module, v.Name) }

var _ Node = (*Verb)(nil)

func (v Verb) schemaChildren() []Node {
	children := make([]Node, 2+len(v.Calls))
	children[0] = v.Request
	children[1] = v.Response
	for i, c := range v.Calls {
		children[i+2] = c
	}
	return children
}
func (v Verb) String() string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "verb %s(%s) %s", v.Name, v.Request, v.Response)
	if len(v.Calls) > 0 {
		fmt.Fprintf(w, "\n  calls %s", v.Calls[0])
		for _, c := range v.Calls[1:] {
			fmt.Fprintf(w, ", %s", c)
		}
	}
	return w.String()
}

var _ Node = (*Module)(nil)

func (m Module) schemaChildren() []Node {
	children := make([]Node, 0, len(m.Data)+len(m.Verbs))
	for _, d := range m.Data {
		children = append(children, d)
	}
	for _, v := range m.Verbs {
		children = append(children, v)
	}
	return children
}
func (m Module) String() string {
	out := &strings.Builder{}
	fmt.Fprintf(out, "module %s {\n", m.Name)
	for _, s := range m.Data {
		fmt.Fprintln(out, indent(s.String()))
	}
	if len(m.Verbs) > 0 {
		fmt.Fprintln(out)
		for _, v := range m.Verbs {
			fmt.Fprintln(out, indent(v.String()))
		}
	}
	fmt.Fprintln(out, "}")
	return out.String()
}

var _ Node = (*Schema)(nil)

func (s Schema) String() string {
	out := &strings.Builder{}
	for i, m := range s.Modules {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprint(out, m)
	}
	return out.String()
}

func (s Schema) schemaChildren() []Node {
	out := make([]Node, len(s.Modules))
	for i, m := range s.Modules {
		out[i] = m
	}
	return out
}

func indent(s string) string {
	return "  " + strings.Join(strings.Split(s, "\n"), "\n  ")
}
