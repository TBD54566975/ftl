package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/errors"
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

func (a Array) MarshalJSON() ([]byte, error) {
	element, err := jsonMarshalType(a.Element)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var inner = struct {
		Type    string          `json:"type"`
		Element json.RawMessage `json:"element"`
	}{"array", element}
	return json.Marshal(inner)
}
func (a Array) schemaChildren() []Node { return []Node{a.Element} }
func (Array) schemaType()              {}
func (a Array) String() string         { return "array<" + a.Element.String() + ">" }
func (a *Array) UnmarshalJSON(b []byte) error {
	var inner struct {
		Element json.RawMessage `json:"element"`
	}
	if err := json.Unmarshal(b, &inner); err != nil {
		return errors.WithStack(err)
	}
	element, err := jsonUnmarshalType(inner.Element)
	if err != nil {
		return errors.WithStack(err)
	}
	a.Element = element
	return nil
}

var _ Type = (*Map)(nil)

func (m Map) MarshalJSON() ([]byte, error) {
	key, err := jsonMarshalType(m.Key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	value, err := jsonMarshalType(m.Value)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var inner = struct {
		Type  string          `json:"type"`
		Key   json.RawMessage `json:"key"`
		Value json.RawMessage `json:"value"`
	}{"array", key, value}
	return json.Marshal(inner)
}
func (m *Map) UnmarshalJSON(b []byte) error {
	var inner struct {
		Key   json.RawMessage `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(b, &inner); err != nil {
		return errors.WithStack(err)
	}
	key, err := jsonUnmarshalType(inner.Key)
	if err != nil {
		return errors.WithStack(err)
	}
	value, err := jsonUnmarshalType(inner.Value)
	if err != nil {
		return errors.WithStack(err)
	}
	m.Key = key
	m.Value = value
	return nil
}
func (m Map) schemaChildren() []Node { return []Node{m.Key, m.Value} }
func (Map) schemaType()              {}
func (m Map) String() string         { return fmt.Sprintf("map<%s, %s>", m.Key.String(), m.Value.String()) }

var _ Node = (*Field)(nil)

func (f Field) MarshalJSON() ([]byte, error) {
	data, err := jsonMarshalType(f.Type)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return json.Marshal(struct {
		Name string          `json:"name"`
		Type json.RawMessage `json:"type"`
	}{f.Name, data})
}
func (f *Field) UnmarshalJSON(b []byte) error {
	var inner struct {
		Name string          `json:"name"`
		Type json.RawMessage `json:"type"`
	}
	if err := json.Unmarshal(b, &inner); err != nil {
		return errors.WithStack(err)
	}
	typ, err := jsonUnmarshalType(inner.Type)
	if err != nil {
		return errors.WithStack(err)
	}
	f.Name = inner.Name
	f.Type = typ
	return nil
}
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
func (s Data) String() string {
	out := &strings.Builder{}
	fmt.Fprintf(out, "data %s {\n", s.Name)
	for _, f := range s.Fields {
		fmt.Fprintln(out, indent(f.String(), "  "))
	}
	fmt.Fprintf(out, "}")
	return out.String()
}

var _ Type = (*VerbRef)(nil)

func (VerbRef) schemaChildren() []Node { return nil }
func (VerbRef) schemaType()            {}
func (v VerbRef) String() string       { return fmt.Sprintf("%s.%s", v.Module, v.Verb) }

var _ Node = (*Verb)(nil)

func (v *Verb) UnmarshalJSON(b []byte) error {
	var inner struct {
		Name     string          `json:"name"`
		Request  json.RawMessage `json:"request"`
		Response json.RawMessage `json:"response"`
		Calls    []VerbRef       `json:"calls,omitempty"`
	}
	if err := json.Unmarshal(b, &inner); err != nil {
		return errors.WithStack(err)
	}
	req, err := jsonUnmarshalType(inner.Request)
	if err != nil {
		return errors.WithStack(err)
	}
	resp, err := jsonUnmarshalType(inner.Response)
	if err != nil {
		return errors.WithStack(err)
	}
	v.Name = inner.Name
	var ok bool
	v.Request, ok = req.(DataRef)
	if !ok {
		return errors.Errorf("verb %q request is not a dataref", v.Name)
	}
	v.Response, ok = resp.(DataRef)
	if !ok {
		return errors.Errorf("verb %q response is not a dataref", v.Name)
	}
	v.Calls = inner.Calls
	return nil
}
func (v Verb) MarshalJSON() ([]byte, error) {
	req, err := jsonMarshalType(v.Request)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := jsonMarshalType(v.Response)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var inner = struct {
		Name     string          `json:"name"`
		Request  json.RawMessage `json:"request"`
		Response json.RawMessage `json:"response"`
		Calls    []VerbRef       `json:"calls,omitempty"`
	}{v.Name, req, resp, v.Calls}
	return json.Marshal(inner)
}
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
		fmt.Fprintln(out, indent(s.String(), "  "))
	}
	if len(m.Verbs) > 0 {
		fmt.Fprintln(out)
		for _, v := range m.Verbs {
			fmt.Fprintln(out, indent(v.String(), "  "))
		}
	}
	fmt.Fprintln(out, "}")
	return out.String()
}

func jsonMarshalType(t Type) ([]byte, error) {
	object := map[string]any{}
	if data, err := json.Marshal(t); err != nil {
		return nil, errors.WithStack(err)
	} else if err = json.Unmarshal(data, &object); err != nil {
		return nil, errors.WithStack(err)
	}
	kind := strings.ToLower(reflect.TypeOf(t).Name())
	object["type"] = kind
	return json.Marshal(object)

}

func jsonUnmarshalType(data []byte) (Type, error) {
	var kind struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &kind); err != nil {
		return nil, errors.WithStack(err)
	}
	var dst Type
	switch kind.Type {
	case "int":
		dst = &Int{}
	case "string":
		dst = &String{}
	case "boolean":
		dst = &Bool{}
	case "float":
		dst = &Float{}
	case "array":
		dst = &Array{}
	case "map":
		dst = &Map{}
	case "dataref":
		dst = &DataRef{}
	case "verbref":
		dst = &VerbRef{}
	default:
		return nil, errors.Errorf("unknown type %q", kind.Type)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return nil, errors.WithStack(err)
	}
	return reflect.ValueOf(dst).Elem().Interface().(Type), nil
}

func indent(s, i string) string {
	return i + strings.Join(strings.Split(s, "\n"), "\n"+i)
}
