// Code generated by go2proto. DO NOT EDIT.

package schema

import "fmt"
import destpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
import "google.golang.org/protobuf/proto"
import "google.golang.org/protobuf/types/known/timestamppb"
import "google.golang.org/protobuf/types/known/durationpb"

var _ fmt.Stringer
var _ = timestamppb.Timestamp{}
var _ = durationpb.Duration{}

// protoSlice converts a slice of values to a slice of protobuf values.
func protoSlice[P any, T interface{ ToProto() P }](values []T) []P {
	out := make([]P, len(values))
	for i, v := range values {
		out[i] = v.ToProto()
	}
	return out
}

// protoSlicef converts a slice of values to a slice of protobuf values using a mapping function.
func protoSlicef[P, T any](values []T, f func(T) P) []P {
	out := make([]P, len(values))
	for i, v := range values {
		out[i] = f(v)
	}
	return out
}

func (x *AWSIAMAuthDatabaseConnector) ToProto() *destpb.AWSIAMAuthDatabaseConnector {
	if x == nil {
		return nil
	}
	return &destpb.AWSIAMAuthDatabaseConnector{
		Pos:      x.Pos.ToProto(),
		Username: string(x.Username),
		Endpoint: string(x.Endpoint),
		Database: string(x.Database),
	}
}

func (x AliasKind) ToProto() destpb.AliasKind {
	return destpb.AliasKind(x)
}

func (x *Any) ToProto() *destpb.Any {
	if x == nil {
		return nil
	}
	return &destpb.Any{
		Pos: x.Pos.ToProto(),
	}
}

func (x *Array) ToProto() *destpb.Array {
	if x == nil {
		return nil
	}
	return &destpb.Array{
		Pos:     x.Pos.ToProto(),
		Element: TypeToProto(x.Element),
	}
}

func (x *Bool) ToProto() *destpb.Bool {
	if x == nil {
		return nil
	}
	return &destpb.Bool{
		Pos: x.Pos.ToProto(),
	}
}

func (x *Bytes) ToProto() *destpb.Bytes {
	if x == nil {
		return nil
	}
	return &destpb.Bytes{
		Pos: x.Pos.ToProto(),
	}
}

func (x *Config) ToProto() *destpb.Config {
	if x == nil {
		return nil
	}
	return &destpb.Config{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Name:     string(x.Name),
		Type:     TypeToProto(x.Type),
	}
}

func (x *DSNDatabaseConnector) ToProto() *destpb.DSNDatabaseConnector {
	if x == nil {
		return nil
	}
	return &destpb.DSNDatabaseConnector{
		Pos: x.Pos.ToProto(),
		Dsn: string(x.DSN),
	}
}

func (x *Data) ToProto() *destpb.Data {
	if x == nil {
		return nil
	}
	return &destpb.Data{
		Pos:            x.Pos.ToProto(),
		Comments:       protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Export:         bool(x.Export),
		Name:           string(x.Name),
		TypeParameters: protoSlice[*destpb.TypeParameter](x.TypeParameters),
		Fields:         protoSlice[*destpb.Field](x.Fields),
		Metadata:       protoSlicef(x.Metadata, MetadataToProto),
	}
}

func (x *Database) ToProto() *destpb.Database {
	if x == nil {
		return nil
	}
	return &destpb.Database{
		Pos:      x.Pos.ToProto(),
		Runtime:  x.Runtime.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Type:     string(x.Type),
		Name:     string(x.Name),
		Metadata: protoSlicef(x.Metadata, MetadataToProto),
	}
}

// DatabaseConnectorToProto converts a DatabaseConnector sum type to a protobuf message.
func DatabaseConnectorToProto(value DatabaseConnector) *destpb.DatabaseConnector {
	switch value := value.(type) {
	case nil:
		return nil
	case *AWSIAMAuthDatabaseConnector:
		return &destpb.DatabaseConnector{
			Value: &destpb.DatabaseConnector_AwsiamAuthDatabaseConnector{value.ToProto()},
		}
	case *DSNDatabaseConnector:
		return &destpb.DatabaseConnector{
			Value: &destpb.DatabaseConnector_DsnDatabaseConnector{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *DatabaseRuntime) ToProto() *destpb.DatabaseRuntime {
	if x == nil {
		return nil
	}
	return &destpb.DatabaseRuntime{
		Connections: x.Connections.ToProto(),
	}
}

func (x *DatabaseRuntimeConnections) ToProto() *destpb.DatabaseRuntimeConnections {
	if x == nil {
		return nil
	}
	return &destpb.DatabaseRuntimeConnections{
		Read:  DatabaseConnectorToProto(x.Read),
		Write: DatabaseConnectorToProto(x.Write),
	}
}

func (x *DatabaseRuntimeConnectionsEvent) ToProto() *destpb.DatabaseRuntimeConnectionsEvent {
	if x == nil {
		return nil
	}
	return &destpb.DatabaseRuntimeConnectionsEvent{
		Connections: x.Connections.ToProto(),
	}
}

func (x *DatabaseRuntimeEvent) ToProto() *destpb.DatabaseRuntimeEvent {
	if x == nil {
		return nil
	}
	return &destpb.DatabaseRuntimeEvent{
		Id:      string(x.ID),
		Payload: DatabaseRuntimeEventPayloadToProto(x.Payload),
	}
}

// DatabaseRuntimeEventPayloadToProto converts a DatabaseRuntimeEventPayload sum type to a protobuf message.
func DatabaseRuntimeEventPayloadToProto(value DatabaseRuntimeEventPayload) *destpb.DatabaseRuntimeEventPayload {
	switch value := value.(type) {
	case nil:
		return nil
	case *DatabaseRuntimeConnectionsEvent:
		return &destpb.DatabaseRuntimeEventPayload{
			Value: &destpb.DatabaseRuntimeEventPayload_DatabaseRuntimeConnectionsEvent{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

// DeclToProto converts a Decl sum type to a protobuf message.
func DeclToProto(value Decl) *destpb.Decl {
	switch value := value.(type) {
	case nil:
		return nil
	case *Config:
		return &destpb.Decl{
			Value: &destpb.Decl_Config{value.ToProto()},
		}
	case *Data:
		return &destpb.Decl{
			Value: &destpb.Decl_Data{value.ToProto()},
		}
	case *Database:
		return &destpb.Decl{
			Value: &destpb.Decl_Database{value.ToProto()},
		}
	case *Enum:
		return &destpb.Decl{
			Value: &destpb.Decl_Enum{value.ToProto()},
		}
	case *Secret:
		return &destpb.Decl{
			Value: &destpb.Decl_Secret{value.ToProto()},
		}
	case *Topic:
		return &destpb.Decl{
			Value: &destpb.Decl_Topic{value.ToProto()},
		}
	case *TypeAlias:
		return &destpb.Decl{
			Value: &destpb.Decl_TypeAlias{value.ToProto()},
		}
	case *Verb:
		return &destpb.Decl{
			Value: &destpb.Decl_Verb{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *Enum) ToProto() *destpb.Enum {
	if x == nil {
		return nil
	}
	return &destpb.Enum{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Export:   bool(x.Export),
		Name:     string(x.Name),
		Type:     TypeToProto(x.Type),
		Variants: protoSlice[*destpb.EnumVariant](x.Variants),
	}
}

func (x *EnumVariant) ToProto() *destpb.EnumVariant {
	if x == nil {
		return nil
	}
	return &destpb.EnumVariant{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Name:     string(x.Name),
		Value:    ValueToProto(x.Value),
	}
}

func (x *Field) ToProto() *destpb.Field {
	if x == nil {
		return nil
	}
	return &destpb.Field{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Name:     string(x.Name),
		Type:     TypeToProto(x.Type),
		Metadata: protoSlicef(x.Metadata, MetadataToProto),
	}
}

func (x *Float) ToProto() *destpb.Float {
	if x == nil {
		return nil
	}
	return &destpb.Float{
		Pos: x.Pos.ToProto(),
	}
}

func (x FromOffset) ToProto() destpb.FromOffset {
	return destpb.FromOffset(x)
}

// IngressPathComponentToProto converts a IngressPathComponent sum type to a protobuf message.
func IngressPathComponentToProto(value IngressPathComponent) *destpb.IngressPathComponent {
	switch value := value.(type) {
	case nil:
		return nil
	case *IngressPathLiteral:
		return &destpb.IngressPathComponent{
			Value: &destpb.IngressPathComponent_IngressPathLiteral{value.ToProto()},
		}
	case *IngressPathParameter:
		return &destpb.IngressPathComponent{
			Value: &destpb.IngressPathComponent_IngressPathParameter{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *IngressPathLiteral) ToProto() *destpb.IngressPathLiteral {
	if x == nil {
		return nil
	}
	return &destpb.IngressPathLiteral{
		Pos:  x.Pos.ToProto(),
		Text: string(x.Text),
	}
}

func (x *IngressPathParameter) ToProto() *destpb.IngressPathParameter {
	if x == nil {
		return nil
	}
	return &destpb.IngressPathParameter{
		Pos:  x.Pos.ToProto(),
		Name: string(x.Name),
	}
}

func (x *Int) ToProto() *destpb.Int {
	if x == nil {
		return nil
	}
	return &destpb.Int{
		Pos: x.Pos.ToProto(),
	}
}

func (x *IntValue) ToProto() *destpb.IntValue {
	if x == nil {
		return nil
	}
	return &destpb.IntValue{
		Pos:   x.Pos.ToProto(),
		Value: int64(x.Value),
	}
}

func (x *Map) ToProto() *destpb.Map {
	if x == nil {
		return nil
	}
	return &destpb.Map{
		Pos:   x.Pos.ToProto(),
		Key:   TypeToProto(x.Key),
		Value: TypeToProto(x.Value),
	}
}

// MetadataToProto converts a Metadata sum type to a protobuf message.
func MetadataToProto(value Metadata) *destpb.Metadata {
	switch value := value.(type) {
	case nil:
		return nil
	case *MetadataAlias:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Alias{value.ToProto()},
		}
	case *MetadataArtefact:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Artefact{value.ToProto()},
		}
	case *MetadataCalls:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Calls{value.ToProto()},
		}
	case *MetadataConfig:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Config{value.ToProto()},
		}
	case *MetadataCronJob:
		return &destpb.Metadata{
			Value: &destpb.Metadata_CronJob{value.ToProto()},
		}
	case *MetadataDatabases:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Databases{value.ToProto()},
		}
	case *MetadataEncoding:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Encoding{value.ToProto()},
		}
	case *MetadataIngress:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Ingress{value.ToProto()},
		}
	case *MetadataPublisher:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Publisher{value.ToProto()},
		}
	case *MetadataRetry:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Retry{value.ToProto()},
		}
	case *MetadataSQLMigration:
		return &destpb.Metadata{
			Value: &destpb.Metadata_SqlMigration{value.ToProto()},
		}
	case *MetadataSecrets:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Secrets{value.ToProto()},
		}
	case *MetadataSubscriber:
		return &destpb.Metadata{
			Value: &destpb.Metadata_Subscriber{value.ToProto()},
		}
	case *MetadataTypeMap:
		return &destpb.Metadata{
			Value: &destpb.Metadata_TypeMap{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *MetadataAlias) ToProto() *destpb.MetadataAlias {
	if x == nil {
		return nil
	}
	return &destpb.MetadataAlias{
		Pos:   x.Pos.ToProto(),
		Kind:  x.Kind.ToProto(),
		Alias: string(x.Alias),
	}
}

func (x *MetadataArtefact) ToProto() *destpb.MetadataArtefact {
	if x == nil {
		return nil
	}
	return &destpb.MetadataArtefact{
		Pos:        x.Pos.ToProto(),
		Path:       string(x.Path),
		Digest:     string(x.Digest),
		Executable: bool(x.Executable),
	}
}

func (x *MetadataCalls) ToProto() *destpb.MetadataCalls {
	if x == nil {
		return nil
	}
	return &destpb.MetadataCalls{
		Pos:   x.Pos.ToProto(),
		Calls: protoSlice[*destpb.Ref](x.Calls),
	}
}

func (x *MetadataConfig) ToProto() *destpb.MetadataConfig {
	if x == nil {
		return nil
	}
	return &destpb.MetadataConfig{
		Pos:    x.Pos.ToProto(),
		Config: protoSlice[*destpb.Ref](x.Config),
	}
}

func (x *MetadataCronJob) ToProto() *destpb.MetadataCronJob {
	if x == nil {
		return nil
	}
	return &destpb.MetadataCronJob{
		Pos:  x.Pos.ToProto(),
		Cron: string(x.Cron),
	}
}

func (x *MetadataDatabases) ToProto() *destpb.MetadataDatabases {
	if x == nil {
		return nil
	}
	return &destpb.MetadataDatabases{
		Pos:   x.Pos.ToProto(),
		Calls: protoSlice[*destpb.Ref](x.Calls),
	}
}

func (x *MetadataEncoding) ToProto() *destpb.MetadataEncoding {
	if x == nil {
		return nil
	}
	return &destpb.MetadataEncoding{
		Pos:     x.Pos.ToProto(),
		Type:    string(x.Type),
		Lenient: bool(x.Lenient),
	}
}

func (x *MetadataIngress) ToProto() *destpb.MetadataIngress {
	if x == nil {
		return nil
	}
	return &destpb.MetadataIngress{
		Pos:    x.Pos.ToProto(),
		Type:   string(x.Type),
		Method: string(x.Method),
		Path:   protoSlicef(x.Path, IngressPathComponentToProto),
	}
}

func (x *MetadataPublisher) ToProto() *destpb.MetadataPublisher {
	if x == nil {
		return nil
	}
	return &destpb.MetadataPublisher{
		Pos:    x.Pos.ToProto(),
		Topics: protoSlice[*destpb.Ref](x.Topics),
	}
}

func (x *MetadataRetry) ToProto() *destpb.MetadataRetry {
	if x == nil {
		return nil
	}
	return &destpb.MetadataRetry{
		Pos:        x.Pos.ToProto(),
		Count:      proto.Int64(int64(*x.Count)),
		MinBackoff: string(x.MinBackoff),
		MaxBackoff: string(x.MaxBackoff),
		Catch:      x.Catch.ToProto(),
	}
}

func (x *MetadataSQLMigration) ToProto() *destpb.MetadataSQLMigration {
	if x == nil {
		return nil
	}
	return &destpb.MetadataSQLMigration{
		Pos:    x.Pos.ToProto(),
		Digest: string(x.Digest),
	}
}

func (x *MetadataSecrets) ToProto() *destpb.MetadataSecrets {
	if x == nil {
		return nil
	}
	return &destpb.MetadataSecrets{
		Pos:     x.Pos.ToProto(),
		Secrets: protoSlice[*destpb.Ref](x.Secrets),
	}
}

func (x *MetadataSubscriber) ToProto() *destpb.MetadataSubscriber {
	if x == nil {
		return nil
	}
	return &destpb.MetadataSubscriber{
		Pos:        x.Pos.ToProto(),
		Topic:      x.Topic.ToProto(),
		FromOffset: x.FromOffset.ToProto(),
		DeadLetter: bool(x.DeadLetter),
	}
}

func (x *MetadataTypeMap) ToProto() *destpb.MetadataTypeMap {
	if x == nil {
		return nil
	}
	return &destpb.MetadataTypeMap{
		Pos:        x.Pos.ToProto(),
		Runtime:    string(x.Runtime),
		NativeName: string(x.NativeName),
	}
}

func (x *Module) ToProto() *destpb.Module {
	if x == nil {
		return nil
	}
	return &destpb.Module{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Builtin:  bool(x.Builtin),
		Name:     string(x.Name),
		Metadata: protoSlicef(x.Metadata, MetadataToProto),
		Decls:    protoSlicef(x.Decls, DeclToProto),
		Runtime:  x.Runtime.ToProto(),
	}
}

func (x *ModuleRuntime) ToProto() *destpb.ModuleRuntime {
	if x == nil {
		return nil
	}
	return &destpb.ModuleRuntime{
		Base:       x.Base.ToProto(),
		Scaling:    x.Scaling.ToProto(),
		Deployment: x.Deployment.ToProto(),
	}
}

func (x *ModuleRuntimeBase) ToProto() *destpb.ModuleRuntimeBase {
	if x == nil {
		return nil
	}
	return &destpb.ModuleRuntimeBase{
		CreateTime: timestamppb.New(x.CreateTime),
		Language:   string(x.Language),
		Os:         proto.String(string(x.OS)),
		Arch:       proto.String(string(x.Arch)),
		Image:      proto.String(string(x.Image)),
	}
}

func (x *ModuleRuntimeDeployment) ToProto() *destpb.ModuleRuntimeDeployment {
	if x == nil {
		return nil
	}
	return &destpb.ModuleRuntimeDeployment{
		Endpoint:      string(x.Endpoint),
		DeploymentKey: string(x.DeploymentKey),
	}
}

// ModuleRuntimeEventToProto converts a ModuleRuntimeEvent sum type to a protobuf message.
func ModuleRuntimeEventToProto(value ModuleRuntimeEvent) *destpb.ModuleRuntimeEvent {
	switch value := value.(type) {
	case nil:
		return nil
	case *ModuleRuntimeBase:
		return &destpb.ModuleRuntimeEvent{
			Value: &destpb.ModuleRuntimeEvent_ModuleRuntimeBase{value.ToProto()},
		}
	case *ModuleRuntimeDeployment:
		return &destpb.ModuleRuntimeEvent{
			Value: &destpb.ModuleRuntimeEvent_ModuleRuntimeDeployment{value.ToProto()},
		}
	case *ModuleRuntimeScaling:
		return &destpb.ModuleRuntimeEvent{
			Value: &destpb.ModuleRuntimeEvent_ModuleRuntimeScaling{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *ModuleRuntimeScaling) ToProto() *destpb.ModuleRuntimeScaling {
	if x == nil {
		return nil
	}
	return &destpb.ModuleRuntimeScaling{
		MinReplicas: int32(x.MinReplicas),
	}
}

func (x *Optional) ToProto() *destpb.Optional {
	if x == nil {
		return nil
	}
	return &destpb.Optional{
		Pos:  x.Pos.ToProto(),
		Type: TypeToProto(x.Type),
	}
}

func (x *Position) ToProto() *destpb.Position {
	if x == nil {
		return nil
	}
	return &destpb.Position{
		Filename: string(x.Filename),
		Line:     int64(x.Line),
		Column:   int64(x.Column),
	}
}

func (x *Ref) ToProto() *destpb.Ref {
	if x == nil {
		return nil
	}
	return &destpb.Ref{
		Pos:            x.Pos.ToProto(),
		Module:         string(x.Module),
		Name:           string(x.Name),
		TypeParameters: protoSlicef(x.TypeParameters, TypeToProto),
	}
}

// RuntimeEventToProto converts a RuntimeEvent sum type to a protobuf message.
func RuntimeEventToProto(value RuntimeEvent) *destpb.RuntimeEvent {
	switch value := value.(type) {
	case nil:
		return nil
	case *DatabaseRuntimeEvent:
		return &destpb.RuntimeEvent{
			Value: &destpb.RuntimeEvent_DatabaseRuntimeEvent{value.ToProto()},
		}
	case *ModuleRuntimeBase:
		return &destpb.RuntimeEvent{
			Value: &destpb.RuntimeEvent_ModuleRuntimeBase{value.ToProto()},
		}
	case *ModuleRuntimeDeployment:
		return &destpb.RuntimeEvent{
			Value: &destpb.RuntimeEvent_ModuleRuntimeDeployment{value.ToProto()},
		}
	case *ModuleRuntimeScaling:
		return &destpb.RuntimeEvent{
			Value: &destpb.RuntimeEvent_ModuleRuntimeScaling{value.ToProto()},
		}
	case *TopicRuntimeEvent:
		return &destpb.RuntimeEvent{
			Value: &destpb.RuntimeEvent_TopicRuntimeEvent{value.ToProto()},
		}
	case *VerbRuntimeEvent:
		return &destpb.RuntimeEvent{
			Value: &destpb.RuntimeEvent_VerbRuntimeEvent{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *Schema) ToProto() *destpb.Schema {
	if x == nil {
		return nil
	}
	return &destpb.Schema{
		Pos:     x.Pos.ToProto(),
		Modules: protoSlice[*destpb.Module](x.Modules),
	}
}

func (x *Secret) ToProto() *destpb.Secret {
	if x == nil {
		return nil
	}
	return &destpb.Secret{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Name:     string(x.Name),
		Type:     TypeToProto(x.Type),
	}
}

func (x *String) ToProto() *destpb.String {
	if x == nil {
		return nil
	}
	return &destpb.String{
		Pos: x.Pos.ToProto(),
	}
}

func (x *StringValue) ToProto() *destpb.StringValue {
	if x == nil {
		return nil
	}
	return &destpb.StringValue{
		Pos:   x.Pos.ToProto(),
		Value: string(x.Value),
	}
}

func (x *Time) ToProto() *destpb.Time {
	if x == nil {
		return nil
	}
	return &destpb.Time{
		Pos: x.Pos.ToProto(),
	}
}

func (x *Topic) ToProto() *destpb.Topic {
	if x == nil {
		return nil
	}
	return &destpb.Topic{
		Pos:      x.Pos.ToProto(),
		Runtime:  x.Runtime.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Export:   bool(x.Export),
		Name:     string(x.Name),
		Event:    TypeToProto(x.Event),
	}
}

func (x *TopicRuntime) ToProto() *destpb.TopicRuntime {
	if x == nil {
		return nil
	}
	return &destpb.TopicRuntime{
		KafkaBrokers: protoSlicef(x.KafkaBrokers, func(v string) string { return string(v) }),
		TopicId:      string(x.TopicID),
	}
}

func (x *TopicRuntimeEvent) ToProto() *destpb.TopicRuntimeEvent {
	if x == nil {
		return nil
	}
	return &destpb.TopicRuntimeEvent{
		Id:      string(x.ID),
		Payload: x.Payload.ToProto(),
	}
}

// TypeToProto converts a Type sum type to a protobuf message.
func TypeToProto(value Type) *destpb.Type {
	switch value := value.(type) {
	case nil:
		return nil
	case *Any:
		return &destpb.Type{
			Value: &destpb.Type_Any{value.ToProto()},
		}
	case *Array:
		return &destpb.Type{
			Value: &destpb.Type_Array{value.ToProto()},
		}
	case *Bool:
		return &destpb.Type{
			Value: &destpb.Type_Bool{value.ToProto()},
		}
	case *Bytes:
		return &destpb.Type{
			Value: &destpb.Type_Bytes{value.ToProto()},
		}
	case *Float:
		return &destpb.Type{
			Value: &destpb.Type_Float{value.ToProto()},
		}
	case *Int:
		return &destpb.Type{
			Value: &destpb.Type_Int{value.ToProto()},
		}
	case *Map:
		return &destpb.Type{
			Value: &destpb.Type_Map{value.ToProto()},
		}
	case *Optional:
		return &destpb.Type{
			Value: &destpb.Type_Optional{value.ToProto()},
		}
	case *Ref:
		return &destpb.Type{
			Value: &destpb.Type_Ref{value.ToProto()},
		}
	case *String:
		return &destpb.Type{
			Value: &destpb.Type_String_{value.ToProto()},
		}
	case *Time:
		return &destpb.Type{
			Value: &destpb.Type_Time{value.ToProto()},
		}
	case *Unit:
		return &destpb.Type{
			Value: &destpb.Type_Unit{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *TypeAlias) ToProto() *destpb.TypeAlias {
	if x == nil {
		return nil
	}
	return &destpb.TypeAlias{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Export:   bool(x.Export),
		Name:     string(x.Name),
		Type:     TypeToProto(x.Type),
		Metadata: protoSlicef(x.Metadata, MetadataToProto),
	}
}

func (x *TypeParameter) ToProto() *destpb.TypeParameter {
	if x == nil {
		return nil
	}
	return &destpb.TypeParameter{
		Pos:  x.Pos.ToProto(),
		Name: string(x.Name),
	}
}

func (x *TypeValue) ToProto() *destpb.TypeValue {
	if x == nil {
		return nil
	}
	return &destpb.TypeValue{
		Pos:   x.Pos.ToProto(),
		Value: TypeToProto(x.Value),
	}
}

func (x *Unit) ToProto() *destpb.Unit {
	if x == nil {
		return nil
	}
	return &destpb.Unit{
		Pos: x.Pos.ToProto(),
	}
}

// ValueToProto converts a Value sum type to a protobuf message.
func ValueToProto(value Value) *destpb.Value {
	switch value := value.(type) {
	case nil:
		return nil
	case *IntValue:
		return &destpb.Value{
			Value: &destpb.Value_IntValue{value.ToProto()},
		}
	case *StringValue:
		return &destpb.Value{
			Value: &destpb.Value_StringValue{value.ToProto()},
		}
	case *TypeValue:
		return &destpb.Value{
			Value: &destpb.Value_TypeValue{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *Verb) ToProto() *destpb.Verb {
	if x == nil {
		return nil
	}
	return &destpb.Verb{
		Pos:      x.Pos.ToProto(),
		Comments: protoSlicef(x.Comments, func(v string) string { return string(v) }),
		Export:   bool(x.Export),
		Name:     string(x.Name),
		Request:  TypeToProto(x.Request),
		Response: TypeToProto(x.Response),
		Metadata: protoSlicef(x.Metadata, MetadataToProto),
		Runtime:  x.Runtime.ToProto(),
	}
}

func (x *VerbRuntime) ToProto() *destpb.VerbRuntime {
	if x == nil {
		return nil
	}
	return &destpb.VerbRuntime{
		Base:         x.Base.ToProto(),
		Subscription: x.Subscription.ToProto(),
	}
}

func (x *VerbRuntimeBase) ToProto() *destpb.VerbRuntimeBase {
	if x == nil {
		return nil
	}
	return &destpb.VerbRuntimeBase{
		CreateTime: timestamppb.New(x.CreateTime),
		StartTime:  timestamppb.New(x.StartTime),
	}
}

func (x *VerbRuntimeEvent) ToProto() *destpb.VerbRuntimeEvent {
	if x == nil {
		return nil
	}
	return &destpb.VerbRuntimeEvent{
		Id:      string(x.ID),
		Payload: VerbRuntimePayloadToProto(x.Payload),
	}
}

// VerbRuntimePayloadToProto converts a VerbRuntimePayload sum type to a protobuf message.
func VerbRuntimePayloadToProto(value VerbRuntimePayload) *destpb.VerbRuntimePayload {
	switch value := value.(type) {
	case nil:
		return nil
	case *VerbRuntimeBase:
		return &destpb.VerbRuntimePayload{
			Value: &destpb.VerbRuntimePayload_VerbRuntimeBase{value.ToProto()},
		}
	case *VerbRuntimeSubscription:
		return &destpb.VerbRuntimePayload{
			Value: &destpb.VerbRuntimePayload_VerbRuntimeSubscription{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *VerbRuntimeSubscription) ToProto() *destpb.VerbRuntimeSubscription {
	if x == nil {
		return nil
	}
	return &destpb.VerbRuntimeSubscription{
		KafkaBrokers: protoSlicef(x.KafkaBrokers, func(v string) string { return string(v) }),
	}
}
