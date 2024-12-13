package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"

	schemapb "github.com/TBD54566975/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/common/slices"
)

//protobuf:2
type Verb struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"2"`
	Export   bool       `parser:"@'export'?" protobuf:"3"`
	Name     string     `parser:"'verb' @Ident" protobuf:"4"`
	Request  Type       `parser:"'(' @@ ')'" protobuf:"5"`
	Response Type       `parser:"@@" protobuf:"6"`
	Metadata []Metadata `parser:"@@*" protobuf:"7"`

	Runtime *VerbRuntime `protobuf:"31634,optional" parser:""`
}

var _ Decl = (*Verb)(nil)
var _ Symbol = (*Verb)(nil)
var _ Provisioned = (*Verb)(nil)

// VerbKind is the kind of Verb: verb, sink, source or empty.
type VerbKind string

const (
	// VerbKindVerb is a normal verb taking an input and an output of any non-unit type.
	VerbKindVerb VerbKind = "verb"
	// VerbKindSink is a verb that takes an input and returns unit.
	VerbKindSink VerbKind = "sink"
	// VerbKindSource is a verb that returns an output and takes unit.
	VerbKindSource VerbKind = "source"
	// VerbKindEmpty is a verb that takes unit and returns unit.
	VerbKindEmpty VerbKind = "empty"
)

// Kind returns the kind of Verb this is.
func (v *Verb) Kind() VerbKind {
	_, inIsUnit := v.Request.(*Unit)
	_, outIsUnit := v.Response.(*Unit)
	switch {
	case inIsUnit && outIsUnit:
		return VerbKindEmpty

	case inIsUnit:
		return VerbKindSource

	case outIsUnit:
		return VerbKindSink

	default:
		return VerbKindVerb
	}
}

func (v *Verb) Position() Position { return v.Pos }

func (v *Verb) schemaDecl()   {}
func (v *Verb) schemaSymbol() {}
func (v *Verb) provisioned()  {}
func (v *Verb) schemaChildren() []Node {
	children := []Node{}
	if v.Request != nil {
		children = append(children, v.Request)
	}
	if v.Response != nil {
		children = append(children, v.Response)
	}
	for _, c := range v.Metadata {
		children = append(children, c)
	}
	return children
}

func (v *Verb) GetName() string  { return v.Name }
func (v *Verb) IsExported() bool { return v.Export }

func (v *Verb) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(v.Comments))
	if v.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "verb %s(%s) %s", v.Name, v.Request, v.Response)
	fmt.Fprint(w, indent(encodeMetadata(v.Metadata)))
	return w.String()
}

// AddCall adds a call reference to the Verb.
func (v *Verb) AddCall(verb *Ref) {
	if c, ok := slices.FindVariant[*MetadataCalls](v.Metadata); ok {
		c.Calls = append(c.Calls, verb)
		return
	}
	v.Metadata = append(v.Metadata, &MetadataCalls{Calls: []*Ref{verb}})
}

// AddConfig adds a config reference to the Verb.
func (v *Verb) AddConfig(config *Ref) {
	if c, ok := slices.FindVariant[*MetadataConfig](v.Metadata); ok {
		c.Config = append(c.Config, config)
		return
	}
	v.Metadata = append(v.Metadata, &MetadataConfig{Config: []*Ref{config}})
}

// AddSecret adds a config reference to the Verb.
func (v *Verb) AddSecret(secret *Ref) {
	if c, ok := slices.FindVariant[*MetadataSecrets](v.Metadata); ok {
		c.Secrets = append(c.Secrets, secret)
		return
	}
	v.Metadata = append(v.Metadata, &MetadataSecrets{Secrets: []*Ref{secret}})
}

// AddDatabase adds a DB reference to the Verb.
func (v *Verb) AddDatabase(db *Ref) {
	if c, ok := slices.FindVariant[*MetadataDatabases](v.Metadata); ok {
		c.Calls = append(c.Calls, db)
		return
	}
	v.Metadata = append(v.Metadata, &MetadataDatabases{Calls: []*Ref{db}})
}

func (v *Verb) AddSubscription(sub *MetadataSubscriber) {
	v.Metadata = append(v.Metadata, sub)
}

// AddTopicPublish adds a topic that this Verb publishes to.
func (v *Verb) AddTopicPublish(topic *Ref) {
	if c, ok := slices.FindVariant[*MetadataPublisher](v.Metadata); ok {
		c.Topics = append(c.Topics, topic)
		return
	}
	v.Metadata = append(v.Metadata, &MetadataPublisher{Topics: []*Ref{topic}})
}

func (v *Verb) SortMetadata() {
	sortMetadata(v.Metadata)
}

func (v *Verb) GetMetadataIngress() optional.Option[*MetadataIngress] {
	if m, ok := slices.FindVariant[*MetadataIngress](v.Metadata); ok {
		return optional.Some(m)
	}
	return optional.None[*MetadataIngress]()
}

func (v *Verb) GetMetadataCronJob() optional.Option[*MetadataCronJob] {
	if m, ok := slices.FindVariant[*MetadataCronJob](v.Metadata); ok {
		return optional.Some(m)
	}
	return optional.None[*MetadataCronJob]()
}

func (v *Verb) GetProvisioned() ResourceSet {
	var result ResourceSet
	for sub := range slices.FilterVariants[*MetadataSubscriber](v.Metadata) {
		result = append(result, &ProvisionedResource{
			Kind: ResourceTypeSubscription,
			Config: &MetadataSubscriber{
				Topic:      sub.Topic,
				FromOffset: sub.FromOffset,
				DeadLetter: sub.DeadLetter,
			},
		})
	}
	return result
}

func (v *Verb) ResourceID() string {
	return v.Name
}

func VerbFromProto(s *schemapb.Verb) *Verb {
	var runtime *VerbRuntime
	if s.Runtime != nil {
		runtime = &VerbRuntime{
			Base: *VerbRuntimeBaseFromProto(s.Runtime.Base),
		}
		if s.Runtime.Base.StartTime != nil {
			runtime.Subscription = VerbRuntimeSubscriptionFromProto(s.Runtime.Subscription)
		}
	}
	return &Verb{
		Pos:      PosFromProto(s.Pos),
		Export:   s.Export,
		Name:     s.Name,
		Comments: s.Comments,
		Request:  TypeFromProto(s.Request),
		Response: TypeFromProto(s.Response),
		Metadata: metadataListToSchema(s.Metadata),
		Runtime:  runtime,
	}
}
