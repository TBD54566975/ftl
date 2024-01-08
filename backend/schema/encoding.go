package schema

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// This file contains the unmarshalling logic as well as support methods for
// visiting and type safety.

var _ Type = (*Optional)(nil)

func (o *Optional) String() string         { return o.Type.String() + "?" }
func (*Optional) schemaType()              {}
func (o *Optional) schemaChildren() []Node { return []Node{o.Type} }

var _ Type = (*Int)(nil)

func (*Int) schemaChildren() []Node { return nil }
func (*Int) schemaType()            {}
func (*Int) String() string         { return "Int" }

var _ Type = (*Float)(nil)

func (*Float) schemaChildren() []Node { return nil }
func (*Float) schemaType()            {}
func (*Float) String() string         { return "Float" }

var _ Type = (*String)(nil)

func (*String) schemaChildren() []Node { return nil }
func (*String) schemaType()            {}
func (*String) String() string         { return "String" }

var _ Type = (*Bytes)(nil)

func (*Bytes) schemaChildren() []Node { return nil }
func (*Bytes) schemaType()            {}
func (*Bytes) String() string         { return "Bytes" }

var _ Type = (*Bool)(nil)

func (*Bool) schemaChildren() []Node { return nil }
func (*Bool) schemaType()            {}
func (*Bool) String() string         { return "Bool" }

var _ Type = (*Time)(nil)

func (*Time) schemaChildren() []Node { return nil }
func (*Time) schemaType()            {}
func (*Time) String() string         { return "Time" }

var _ Type = (*Array)(nil)

func (a *Array) schemaChildren() []Node { return []Node{a.Element} }
func (*Array) schemaType()              {}
func (a *Array) String() string         { return "[" + a.Element.String() + "]" }

var _ Type = (*Map)(nil)

func (m *Map) schemaChildren() []Node { return []Node{m.Key, m.Value} }
func (*Map) schemaType()              {}
func (m *Map) String() string         { return fmt.Sprintf("{%s: %s}", m.Key.String(), m.Value.String()) }

var _ Node = (*Field)(nil)

func (f *Field) schemaChildren() []Node { return []Node{f.Type} }
func (f *Field) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(f.Comments))
	fmt.Fprintf(w, "%s %s", f.Name, f.Type.String())
	return w.String()
}

var _ Type = (*DataRef)(nil)

func (*DataRef) schemaChildren() []Node { return nil }
func (*DataRef) schemaType()            {}
func (s DataRef) String() string        { return makeRef(s.Module, s.Name) }

var _ Decl = (*Data)(nil)

// schemaDecl implements Decl
func (*Data) schemaDecl() {}
func (d *Data) schemaChildren() []Node {
	children := make([]Node, 0, len(d.Fields)+len(d.Metadata))
	for _, f := range d.Fields {
		children = append(children, f)
	}
	for _, c := range d.Metadata {
		children = append(children, c)
	}
	return children
}
func (d *Data) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(d.Comments))
	fmt.Fprintf(w, "data %s {\n", d.Name)
	for _, f := range d.Fields {
		fmt.Fprintln(w, indent(f.String()))
	}
	fmt.Fprintf(w, "}")
	fmt.Fprint(w, indent(encodeMetadata(d.Metadata)))
	return w.String()
}

var _ Type = (*VerbRef)(nil)

func (*VerbRef) schemaChildren() []Node { return nil }
func (*VerbRef) schemaType()            {}
func (v VerbRef) String() string        { return makeRef(v.Module, v.Name) }

var _ Decl = (*Verb)(nil)

func (v *Verb) schemaDecl() {}
func (v *Verb) schemaChildren() []Node {
	children := make([]Node, 2+len(v.Metadata))
	children[0] = v.Request
	children[1] = v.Response
	for i, c := range v.Metadata {
		children[i+2] = c
	}
	return children
}
func (v *Verb) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(v.Comments))
	fmt.Fprintf(w, "verb %s(%s) %s", v.Name, v.Request, v.Response)
	fmt.Fprint(w, indent(encodeMetadata(v.Metadata)))
	return w.String()
}

var _ Metadata = (*MetadataCalls)(nil)

func (m *MetadataCalls) String() string {
	out := &strings.Builder{}
	fmt.Fprint(out, "calls ")
	w := 6
	for i, call := range m.Calls {
		if i > 0 {
			fmt.Fprint(out, ", ")
			w += 2
		}
		str := call.String()
		if w+len(str) > 70 {
			w = 6
			fmt.Fprint(out, "\n      ")
		}
		w += len(str)
		fmt.Fprint(out, str)
	}
	fmt.Fprintln(out)
	return out.String()
}

func (m *MetadataCalls) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Calls))
	for _, ref := range m.Calls {
		out = append(out, ref)
	}
	return out
}
func (*MetadataCalls) schemaMetadata() {}

var _ Metadata = (*MetadataIngress)(nil)

func (m *MetadataIngress) String() string {
	path := make([]string, len(m.Path))
	for i, p := range m.Path {
		switch v := p.(type) {
		case *IngressPathLiteral:
			path[i] = v.Text
		case *IngressPathParameter:
			path[i] = fmt.Sprintf("{%s}", v.Name)
		}
	}
	return fmt.Sprintf("ingress %s /%s", strings.ToUpper(m.Method), strings.Join(path, "/"))
}

func (m *MetadataIngress) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Path))
	for _, ref := range m.Path {
		out = append(out, ref)
	}
	return out
}

func (*MetadataIngress) schemaMetadata() {}

var _ IngressPathComponent = (*IngressPathLiteral)(nil)

func (l *IngressPathLiteral) String() string            { return l.Text }
func (*IngressPathLiteral) schemaChildren() []Node      { return nil }
func (*IngressPathLiteral) schemaIngressPathComponent() {}

var _ IngressPathComponent = (*IngressPathParameter)(nil)

func (l *IngressPathParameter) String() string            { return l.Name }
func (*IngressPathParameter) schemaChildren() []Node      { return nil }
func (*IngressPathParameter) schemaIngressPathComponent() {}

var _ Node = (*Module)(nil)

func (m *Module) schemaChildren() []Node {
	children := make([]Node, 0, len(m.Decls))
	for _, d := range m.Decls {
		children = append(children, d)
	}
	return children
}
func (m *Module) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(m.Comments))
	fmt.Fprintf(w, "module %s {\n", m.Name)
	for i, s := range m.Decls {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, indent(s.String()))
	}
	fmt.Fprintln(w, "}")
	return w.String()
}

var _ Node = (*Schema)(nil)

func (s *Schema) String() string {
	out := &strings.Builder{}
	for i, m := range s.Modules {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprint(out, m)
	}
	return out.String()
}

func (s *Schema) schemaChildren() []Node {
	out := make([]Node, len(s.Modules))
	for i, m := range s.Modules {
		out[i] = m
	}
	return out
}

func (s *Schema) Hash() [sha256.Size]byte {
	return sha256.Sum256([]byte(s.String()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "  " + strings.Join(strings.Split(s, "\n"), "\n  ")
}

func encodeMetadata(metadata []Metadata) string {
	if len(metadata) == 0 {
		return ""
	}
	w := &strings.Builder{}
	fmt.Fprintln(w)
	for _, c := range metadata {
		fmt.Fprint(w, indent(c.String()))
	}
	return w.String()
}

func encodeComments(comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	w := &strings.Builder{}
	for _, c := range comments {
		fmt.Fprintf(w, "// %s\n", c)
	}
	return w.String()
}

func makeRef(module, name string) string {
	if module == "" {
		return name
	}
	return module + "." + name
}
