package schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

type DatabaseRuntime struct {
	Connections *DatabaseRuntimeConnections `parser:"" protobuf:"1,optional"`
}

var _ Symbol = (*DatabaseRuntime)(nil)

func (d *DatabaseRuntime) Position() Position { return d.Connections.Read.Position() }
func (d *DatabaseRuntime) schemaSymbol()      {}
func (d *DatabaseRuntime) String() string {
	return fmt.Sprintf("read: %s, write: %s", d.Connections.Read, d.Connections.Write)
}
func (d *DatabaseRuntime) ToProto() protoreflect.ProtoMessage {
	rc, ok := d.Connections.Read.ToProto().(*schemapb.DatabaseConnector)
	if !ok {
		panic(fmt.Sprintf("unknown database connector type: %T", d.Connections.Read))
	}
	wc, ok := d.Connections.Write.ToProto().(*schemapb.DatabaseConnector)
	if !ok {
		panic(fmt.Sprintf("unknown database connector type: %T", d.Connections.Write))
	}

	return &schemapb.DatabaseRuntime{
		Connections: &schemapb.DatabaseRuntimeConnections{
			Read:  rc,
			Write: wc,
		},
	}
}
func (d *DatabaseRuntime) schemaChildren() []Node {
	return []Node{d.Connections}
}

func (d *DatabaseRuntime) ApplyEvent(e *DatabaseRuntimeEvent) {
	switch e := e.Payload.(type) {
	case *DatabaseRuntimeConnectionsEvent:
		d.Connections = e.Connections
	default:
		panic(fmt.Sprintf("unknown database runtime event type: %T", e))
	}
}

type DatabaseRuntimeConnections struct {
	Read  DatabaseConnector `parser:"" protobuf:"1"`
	Write DatabaseConnector `parser:"" protobuf:"2"`
}

var _ Symbol = (*DatabaseRuntimeConnections)(nil)

func (d *DatabaseRuntimeConnections) Position() Position { return d.Read.Position() }
func (d *DatabaseRuntimeConnections) schemaSymbol()      {}
func (d *DatabaseRuntimeConnections) String() string {
	return fmt.Sprintf("read: %s, write: %s", d.Read, d.Write)
}
func (d *DatabaseRuntimeConnections) ToProto() protoreflect.ProtoMessage {
	return &schemapb.DatabaseRuntimeConnections{
		Read:  d.Read.ToProto().(*schemapb.DatabaseConnector),  //nolint:forcetypeassert
		Write: d.Write.ToProto().(*schemapb.DatabaseConnector), //nolint:forcetypeassert
	}
}

func DatabaseRuntimeConnectionsFromProto(s *schemapb.DatabaseRuntimeConnections) *DatabaseRuntimeConnections {
	return &DatabaseRuntimeConnections{
		Read:  DatabaseConnectorFromProto(s.Read),
		Write: DatabaseConnectorFromProto(s.Write),
	}
}

func (d *DatabaseRuntimeConnections) schemaChildren() []Node {
	return []Node{d.Read, d.Write}
}

type DatabaseConnector interface {
	Node

	databaseConnector()
}

//protobuf:1
type DSNDatabaseConnector struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	DSN string `parser:"" protobuf:"2"`
}

var _ DatabaseConnector = (*DSNDatabaseConnector)(nil)

func (d *DSNDatabaseConnector) Position() Position { return d.Pos }
func (d *DSNDatabaseConnector) databaseConnector() {}
func (d *DSNDatabaseConnector) String() string     { return d.DSN }
func (d *DSNDatabaseConnector) ToProto() protoreflect.ProtoMessage {
	if d == nil {
		return nil
	}
	return &schemapb.DatabaseConnector{
		Value: &schemapb.DatabaseConnector_DsnDatabaseConnector{
			DsnDatabaseConnector: &schemapb.DSNDatabaseConnector{Dsn: d.DSN},
		},
	}
}

func (d *DSNDatabaseConnector) schemaChildren() []Node { return nil }

//protobuf:2
type AWSIAMAuthDatabaseConnector struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Username string `parser:"" protobuf:"2"`
	Endpoint string `parser:"" protobuf:"3"`
	Database string `parser:"" protobuf:"4"`
}

var _ DatabaseConnector = (*AWSIAMAuthDatabaseConnector)(nil)

func (d *AWSIAMAuthDatabaseConnector) Position() Position { return d.Pos }
func (d *AWSIAMAuthDatabaseConnector) databaseConnector() {}
func (d *AWSIAMAuthDatabaseConnector) String() string {
	return fmt.Sprintf("%s@%s/%s", d.Username, d.Endpoint, d.Database)
}
func (d *AWSIAMAuthDatabaseConnector) ToProto() protoreflect.ProtoMessage {
	if d == nil {
		return nil
	}
	return &schemapb.DatabaseConnector{
		Value: &schemapb.DatabaseConnector_AwsiamAuthDatabaseConnector{
			AwsiamAuthDatabaseConnector: &schemapb.AWSIAMAuthDatabaseConnector{
				Username: d.Username,
				Endpoint: d.Endpoint,
				Database: d.Database,
			},
		},
	}
}

func (d *AWSIAMAuthDatabaseConnector) schemaChildren() []Node { return nil }

func DatabaseRuntimeFromProto(s *schemapb.DatabaseRuntime) *DatabaseRuntime {
	if s == nil {
		return nil
	}
	if s.Connections == nil {
		return &DatabaseRuntime{}
	}
	return &DatabaseRuntime{
		Connections: &DatabaseRuntimeConnections{
			Read:  DatabaseConnectorFromProto(s.Connections.Read),
			Write: DatabaseConnectorFromProto(s.Connections.Write),
		},
	}
}

func DatabaseConnectorFromProto(s *schemapb.DatabaseConnector) DatabaseConnector {
	switch s := s.Value.(type) {
	case *schemapb.DatabaseConnector_DsnDatabaseConnector:
		return &DSNDatabaseConnector{DSN: s.DsnDatabaseConnector.Dsn}
	case *schemapb.DatabaseConnector_AwsiamAuthDatabaseConnector:
		return &AWSIAMAuthDatabaseConnector{
			Username: s.AwsiamAuthDatabaseConnector.Username,
			Endpoint: s.AwsiamAuthDatabaseConnector.Endpoint,
			Database: s.AwsiamAuthDatabaseConnector.Database,
		}
	default:
		panic(fmt.Sprintf("unknown database connector type: %T", s))
	}
}

type DatabaseRuntimeEvent struct {
	ID      string                      `parser:"" protobuf:"1"`
	Payload DatabaseRuntimeEventPayload `parser:"" protobuf:"2"`
}

func (d *DatabaseRuntimeEvent) ApplyTo(s *Module) {
	for _, decl := range s.Decls {
		if db, ok := decl.(*Database); ok && db.Name == d.ID {
			if db.Runtime == nil {
				db.Runtime = &DatabaseRuntime{}
			}
			db.Runtime.ApplyEvent(d)
		}
	}
}

func (d *DatabaseRuntimeEvent) ToProto() *schemapb.DatabaseRuntimeEvent {
	return &schemapb.DatabaseRuntimeEvent{
		Id:      d.ID,
		Payload: d.Payload.ToProto(),
	}
}

func DatabaseRuntimeEventFromProto(s *schemapb.DatabaseRuntimeEvent) *DatabaseRuntimeEvent {
	return &DatabaseRuntimeEvent{
		ID:      s.Id,
		Payload: DatabaseRuntimeEventPayloadFromProto(s.Payload),
	}
}

//sumtype:decl
type DatabaseRuntimeEventPayload interface {
	ToProto() *schemapb.DatabaseRuntimeEventPayload

	databaseRuntimeEventPayload()
}

func DatabaseRuntimeEventPayloadFromProto(s *schemapb.DatabaseRuntimeEventPayload) DatabaseRuntimeEventPayload {
	switch s := s.Value.(type) {
	case *schemapb.DatabaseRuntimeEventPayload_DatabaseRuntimeConnectionsEvent:
		return DatabaseRuntimeConnectionsEventFromProto(s.DatabaseRuntimeConnectionsEvent)
	default:
		panic(fmt.Sprintf("unknown database runtime event payload type: %T", s))
	}
}

//protobuf:1
type DatabaseRuntimeConnectionsEvent struct {
	Connections *DatabaseRuntimeConnections `parser:"" protobuf:"1"`
}

var _ DatabaseRuntimeEventPayload = (*DatabaseRuntimeConnectionsEvent)(nil)

func (d *DatabaseRuntimeConnectionsEvent) databaseRuntimeEventPayload() {}

func DatabaseRuntimeConnectionsEventFromProto(s *schemapb.DatabaseRuntimeConnectionsEvent) *DatabaseRuntimeConnectionsEvent {
	if s == nil {
		return nil
	}
	return &DatabaseRuntimeConnectionsEvent{
		Connections: DatabaseRuntimeConnectionsFromProto(s.Connections),
	}
}

func (d *DatabaseRuntimeConnectionsEvent) ToProto() *schemapb.DatabaseRuntimeEventPayload {
	return &schemapb.DatabaseRuntimeEventPayload{
		Value: &schemapb.DatabaseRuntimeEventPayload_DatabaseRuntimeConnectionsEvent{
			DatabaseRuntimeConnectionsEvent: &schemapb.DatabaseRuntimeConnectionsEvent{
				Connections: d.Connections.ToProto().(*schemapb.DatabaseRuntimeConnections), //nolint:forcetypeassert
			},
		},
	}
}
