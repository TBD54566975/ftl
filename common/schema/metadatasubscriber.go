package schema

import (
	"fmt"
	"strings"
)

type FromOffset int

const (
	FromOffsetUnspecified FromOffset = iota
	FromOffsetBeginning
	FromOffsetLatest
)

func (o *FromOffset) Capture(values []string) error {
	switch values[0] {
	case "beginning":
		*o = FromOffsetBeginning
	case "latest":
		*o = FromOffsetLatest
	default:
		return fmt.Errorf("unexpected value %q", values[0])
	}
	return nil
}

func (o FromOffset) String() string {
	switch o {
	case FromOffsetBeginning:
		return "beginning"
	case FromOffsetLatest:
		return "latest"
	default:
		panic(fmt.Sprintf("unexpected value %d", o))
	}
}

//protobuf:7
type MetadataSubscriber struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Topic      *Ref       `parser:"'+' 'subscribe' @@" protobuf:"2"`
	FromOffset FromOffset `parser:"'from' '=' @('beginning'|'latest')" protobuf:"3"`
	DeadLetter bool       `parser:"@'deadletter'?" protobuf:"4"`
}

var _ Metadata = (*MetadataSubscriber)(nil)

func (*MetadataSubscriber) schemaMetadata() {}
func (m *MetadataSubscriber) schemaChildren() []Node {
	if m.Topic == nil {
		return nil
	}
	return []Node{m.Topic}
}
func (m *MetadataSubscriber) Position() Position { return m.Pos }
func (m *MetadataSubscriber) String() string {
	components := []string{
		"+subscribe",
		m.Topic.String(),
	}
	components = append(components, "from="+m.FromOffset.String())
	if m.DeadLetter {
		components = append(components, "deadletter")
	}
	return strings.Join(components, " ")
}

func DeadLetterNameForSubscriber(verb string) string {
	return verb + "Failed"
}
