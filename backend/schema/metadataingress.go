package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type MetadataIngress struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Type   string                 `parser:"'ingress' @('http' | 'ftl')?" protobuf:"2"`
	Method string                 `parser:"@('GET' | 'POST' | 'PUT' | 'DELETE')" protobuf:"3"`
	Path   []IngressPathComponent `parser:"('/' @@)+" protobuf:"4"`
}

var _ Metadata = (*MetadataIngress)(nil)

func (m *MetadataIngress) Position() Position { return m.Pos }
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
	return fmt.Sprintf("ingress %s %s /%s", m.Type, strings.ToUpper(m.Method), strings.Join(path, "/"))
}

func (m *MetadataIngress) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Path))
	for _, ref := range m.Path {
		out = append(out, ref)
	}
	return out
}

func (*MetadataIngress) schemaMetadata() {}

func (m *MetadataIngress) ToProto() proto.Message {
	return &schemapb.MetadataIngress{
		Pos:    posToProto(m.Pos),
		Type:   m.Type,
		Method: m.Method,
		Path:   ingressListToProto(m.Path),
	}
}
func ingressPathComponentListToSchema(s []*schemapb.IngressPathComponent) []IngressPathComponent {
	var out []IngressPathComponent
	for _, n := range s {
		switch n := n.Value.(type) {
		case *schemapb.IngressPathComponent_IngressPathLiteral:
			out = append(out, &IngressPathLiteral{
				Pos:  posFromProto(n.IngressPathLiteral.Pos),
				Text: n.IngressPathLiteral.Text,
			})
		case *schemapb.IngressPathComponent_IngressPathParameter:
			out = append(out, &IngressPathParameter{
				Pos:  posFromProto(n.IngressPathParameter.Pos),
				Name: n.IngressPathParameter.Name,
			})
		}
	}

	return out
}

type IngressPathComponent interface {
	Node
	schemaIngressPathComponent()
}

type IngressPathLiteral struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Text string `parser:"@~(Whitespace | '/' | '{' | '}')+" protobuf:"2"`
}

var _ IngressPathComponent = (*IngressPathLiteral)(nil)

func (l *IngressPathLiteral) Position() Position        { return l.Pos }
func (l *IngressPathLiteral) String() string            { return l.Text }
func (*IngressPathLiteral) schemaChildren() []Node      { return nil }
func (*IngressPathLiteral) schemaIngressPathComponent() {}
func (l *IngressPathLiteral) ToProto() proto.Message {
	return &schemapb.IngressPathLiteral{Text: l.Text}
}

type IngressPathParameter struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"'{' @Ident '}'" protobuf:"2"`
}

var _ IngressPathComponent = (*IngressPathParameter)(nil)

func (l *IngressPathParameter) Position() Position        { return l.Pos }
func (l *IngressPathParameter) String() string            { return l.Name }
func (*IngressPathParameter) schemaChildren() []Node      { return nil }
func (*IngressPathParameter) schemaIngressPathComponent() {}
func (l *IngressPathParameter) ToProto() proto.Message {
	return &schemapb.IngressPathParameter{Name: l.Name}
}
