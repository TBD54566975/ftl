// THIS FILE IS GENERATED; DO NOT MODIFY

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: model.proto

package testdatapb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Enum int32

const (
	Enum_ENUM_A Enum = 0
	Enum_ENUM_B Enum = 1
)

// Enum value maps for Enum.
var (
	Enum_name = map[int32]string{
		0: "ENUM_A",
		1: "ENUM_B",
	}
	Enum_value = map[string]int32{
		"ENUM_A": 0,
		"ENUM_B": 1,
	}
)

func (x Enum) Enum() *Enum {
	p := new(Enum)
	*p = x
	return p
}

func (x Enum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Enum) Descriptor() protoreflect.EnumDescriptor {
	return file_model_proto_enumTypes[0].Descriptor()
}

func (Enum) Type() protoreflect.EnumType {
	return &file_model_proto_enumTypes[0]
}

func (x Enum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Enum.Descriptor instead.
func (Enum) EnumDescriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{0}
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time     *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
	Duration *durationpb.Duration   `protobuf:"bytes,2,opt,name=duration,proto3" json:"duration,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	mi := &file_model_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetTime() *timestamppb.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

func (x *Message) GetDuration() *durationpb.Duration {
	if x != nil {
		return x.Duration
	}
	return nil
}

type Root struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Int            int64      `protobuf:"varint,1,opt,name=int,proto3" json:"int,omitempty"`
	String_        string     `protobuf:"bytes,2,opt,name=string,proto3" json:"string,omitempty"`
	MessagePtr     *Message   `protobuf:"bytes,4,opt,name=message_ptr,json=messagePtr,proto3" json:"message_ptr,omitempty"`
	Enum           Enum       `protobuf:"varint,5,opt,name=enum,proto3,enum=xyz.block.ftl.go2proto.test.Enum" json:"enum,omitempty"`
	SumType        *SumType   `protobuf:"bytes,6,opt,name=sum_type,json=sumType,proto3" json:"sum_type,omitempty"`
	OptionalInt    *int64     `protobuf:"varint,7,opt,name=optional_int,json=optionalInt,proto3,oneof" json:"optional_int,omitempty"`
	OptionalIntPtr *int64     `protobuf:"varint,8,opt,name=optional_int_ptr,json=optionalIntPtr,proto3,oneof" json:"optional_int_ptr,omitempty"`
	OptionalMsg    *Message   `protobuf:"bytes,9,opt,name=optional_msg,json=optionalMsg,proto3,oneof" json:"optional_msg,omitempty"`
	RepeatedInt    []int64    `protobuf:"varint,10,rep,packed,name=repeated_int,json=repeatedInt,proto3" json:"repeated_int,omitempty"`
	RepeatedMsg    []*Message `protobuf:"bytes,11,rep,name=repeated_msg,json=repeatedMsg,proto3" json:"repeated_msg,omitempty"`
}

func (x *Root) Reset() {
	*x = Root{}
	mi := &file_model_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Root) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Root) ProtoMessage() {}

func (x *Root) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Root.ProtoReflect.Descriptor instead.
func (*Root) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{1}
}

func (x *Root) GetInt() int64 {
	if x != nil {
		return x.Int
	}
	return 0
}

func (x *Root) GetString_() string {
	if x != nil {
		return x.String_
	}
	return ""
}

func (x *Root) GetMessagePtr() *Message {
	if x != nil {
		return x.MessagePtr
	}
	return nil
}

func (x *Root) GetEnum() Enum {
	if x != nil {
		return x.Enum
	}
	return Enum_ENUM_A
}

func (x *Root) GetSumType() *SumType {
	if x != nil {
		return x.SumType
	}
	return nil
}

func (x *Root) GetOptionalInt() int64 {
	if x != nil && x.OptionalInt != nil {
		return *x.OptionalInt
	}
	return 0
}

func (x *Root) GetOptionalIntPtr() int64 {
	if x != nil && x.OptionalIntPtr != nil {
		return *x.OptionalIntPtr
	}
	return 0
}

func (x *Root) GetOptionalMsg() *Message {
	if x != nil {
		return x.OptionalMsg
	}
	return nil
}

func (x *Root) GetRepeatedInt() []int64 {
	if x != nil {
		return x.RepeatedInt
	}
	return nil
}

func (x *Root) GetRepeatedMsg() []*Message {
	if x != nil {
		return x.RepeatedMsg
	}
	return nil
}

type SumType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//
	//	*SumType_A
	//	*SumType_B
	//	*SumType_C
	Value isSumType_Value `protobuf_oneof:"value"`
}

func (x *SumType) Reset() {
	*x = SumType{}
	mi := &file_model_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SumType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SumType) ProtoMessage() {}

func (x *SumType) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SumType.ProtoReflect.Descriptor instead.
func (*SumType) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{2}
}

func (m *SumType) GetValue() isSumType_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *SumType) GetA() *SumTypeA {
	if x, ok := x.GetValue().(*SumType_A); ok {
		return x.A
	}
	return nil
}

func (x *SumType) GetB() *SumTypeB {
	if x, ok := x.GetValue().(*SumType_B); ok {
		return x.B
	}
	return nil
}

func (x *SumType) GetC() *SumTypeC {
	if x, ok := x.GetValue().(*SumType_C); ok {
		return x.C
	}
	return nil
}

type isSumType_Value interface {
	isSumType_Value()
}

type SumType_A struct {
	A *SumTypeA `protobuf:"bytes,1,opt,name=a,proto3,oneof"`
}

type SumType_B struct {
	B *SumTypeB `protobuf:"bytes,2,opt,name=b,proto3,oneof"`
}

type SumType_C struct {
	C *SumTypeC `protobuf:"bytes,3,opt,name=c,proto3,oneof"`
}

func (*SumType_A) isSumType_Value() {}

func (*SumType_B) isSumType_Value() {}

func (*SumType_C) isSumType_Value() {}

type SumTypeA struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	A string `protobuf:"bytes,1,opt,name=a,proto3" json:"a,omitempty"`
}

func (x *SumTypeA) Reset() {
	*x = SumTypeA{}
	mi := &file_model_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SumTypeA) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SumTypeA) ProtoMessage() {}

func (x *SumTypeA) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SumTypeA.ProtoReflect.Descriptor instead.
func (*SumTypeA) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{3}
}

func (x *SumTypeA) GetA() string {
	if x != nil {
		return x.A
	}
	return ""
}

type SumTypeB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	B int64 `protobuf:"varint,1,opt,name=b,proto3" json:"b,omitempty"`
}

func (x *SumTypeB) Reset() {
	*x = SumTypeB{}
	mi := &file_model_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SumTypeB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SumTypeB) ProtoMessage() {}

func (x *SumTypeB) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SumTypeB.ProtoReflect.Descriptor instead.
func (*SumTypeB) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{4}
}

func (x *SumTypeB) GetB() int64 {
	if x != nil {
		return x.B
	}
	return 0
}

type SumTypeC struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	C float64 `protobuf:"fixed64,1,opt,name=c,proto3" json:"c,omitempty"`
}

func (x *SumTypeC) Reset() {
	*x = SumTypeC{}
	mi := &file_model_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SumTypeC) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SumTypeC) ProtoMessage() {}

func (x *SumTypeC) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SumTypeC.ProtoReflect.Descriptor instead.
func (*SumTypeC) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{5}
}

func (x *SumTypeC) GetC() float64 {
	if x != nil {
		return x.C
	}
	return 0
}

var File_model_proto protoreflect.FileDescriptor

var file_model_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67, 0x6f, 0x32,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x70, 0x0a, 0x07, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xb7, 0x04,
	0x0a, 0x04, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x03, 0x69, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67,
	0x12, 0x45, 0x0a, 0x0b, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x70, 0x74, 0x72, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67, 0x6f, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x0a, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x50, 0x74, 0x72, 0x12, 0x35, 0x0a, 0x04, 0x65, 0x6e, 0x75, 0x6d, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67, 0x6f, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x2e, 0x45, 0x6e, 0x75, 0x6d, 0x52, 0x04, 0x65, 0x6e, 0x75, 0x6d, 0x12, 0x3f,
	0x0a, 0x08, 0x73, 0x75, 0x6d, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x67, 0x6f, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x53,
	0x75, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x52, 0x07, 0x73, 0x75, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x26, 0x0a, 0x0c, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x69, 0x6e, 0x74, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x0b, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61,
	0x6c, 0x49, 0x6e, 0x74, 0x88, 0x01, 0x01, 0x12, 0x2d, 0x0a, 0x10, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x61, 0x6c, 0x5f, 0x69, 0x6e, 0x74, 0x5f, 0x70, 0x74, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x03, 0x48, 0x01, 0x52, 0x0e, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x49, 0x6e, 0x74,
	0x50, 0x74, 0x72, 0x88, 0x01, 0x01, 0x12, 0x4c, 0x0a, 0x0c, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x61, 0x6c, 0x5f, 0x6d, 0x73, 0x67, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67, 0x6f, 0x32,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x48, 0x02, 0x52, 0x0b, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x4d, 0x73,
	0x67, 0x88, 0x01, 0x01, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x69, 0x6e, 0x74, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0b, 0x72, 0x65, 0x70, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x49, 0x6e, 0x74, 0x12, 0x47, 0x0a, 0x0c, 0x72, 0x65, 0x70, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x5f, 0x6d, 0x73, 0x67, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67, 0x6f,
	0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x52, 0x0b, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64, 0x4d, 0x73, 0x67,
	0x42, 0x0f, 0x0a, 0x0d, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x69, 0x6e,
	0x74, 0x42, 0x13, 0x0a, 0x11, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x69,
	0x6e, 0x74, 0x5f, 0x70, 0x74, 0x72, 0x42, 0x0f, 0x0a, 0x0d, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x61, 0x6c, 0x5f, 0x6d, 0x73, 0x67, 0x22, 0xb7, 0x01, 0x0a, 0x07, 0x53, 0x75, 0x6d, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x35, 0x0a, 0x01, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67,
	0x6f, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x75, 0x6d,
	0x54, 0x79, 0x70, 0x65, 0x41, 0x48, 0x00, 0x52, 0x01, 0x61, 0x12, 0x35, 0x0a, 0x01, 0x62, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67, 0x6f, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x2e, 0x53, 0x75, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x42, 0x48, 0x00, 0x52, 0x01,
	0x62, 0x12, 0x35, 0x0a, 0x01, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x67, 0x6f, 0x32,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x75, 0x6d, 0x54, 0x79,
	0x70, 0x65, 0x43, 0x48, 0x00, 0x52, 0x01, 0x63, 0x42, 0x07, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x22, 0x18, 0x0a, 0x08, 0x53, 0x75, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x41, 0x12, 0x0c, 0x0a,
	0x01, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x01, 0x61, 0x22, 0x18, 0x0a, 0x08, 0x53,
	0x75, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x42, 0x12, 0x0c, 0x0a, 0x01, 0x62, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x01, 0x62, 0x22, 0x18, 0x0a, 0x08, 0x53, 0x75, 0x6d, 0x54, 0x79, 0x70, 0x65,
	0x43, 0x12, 0x0c, 0x0a, 0x01, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x63, 0x2a,
	0x1e, 0x0a, 0x04, 0x45, 0x6e, 0x75, 0x6d, 0x12, 0x0a, 0x0a, 0x06, 0x45, 0x4e, 0x55, 0x4d, 0x5f,
	0x41, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x42, 0x10, 0x01, 0x42,
	0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x42,
	0x44, 0x35, 0x34, 0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x63, 0x6d,
	0x64, 0x2f, 0x67, 0x6f, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x64,
	0x61, 0x74, 0x61, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_model_proto_rawDescOnce sync.Once
	file_model_proto_rawDescData = file_model_proto_rawDesc
)

func file_model_proto_rawDescGZIP() []byte {
	file_model_proto_rawDescOnce.Do(func() {
		file_model_proto_rawDescData = protoimpl.X.CompressGZIP(file_model_proto_rawDescData)
	})
	return file_model_proto_rawDescData
}

var file_model_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_model_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_model_proto_goTypes = []any{
	(Enum)(0),                     // 0: xyz.block.ftl.go2proto.test.Enum
	(*Message)(nil),               // 1: xyz.block.ftl.go2proto.test.Message
	(*Root)(nil),                  // 2: xyz.block.ftl.go2proto.test.Root
	(*SumType)(nil),               // 3: xyz.block.ftl.go2proto.test.SumType
	(*SumTypeA)(nil),              // 4: xyz.block.ftl.go2proto.test.SumTypeA
	(*SumTypeB)(nil),              // 5: xyz.block.ftl.go2proto.test.SumTypeB
	(*SumTypeC)(nil),              // 6: xyz.block.ftl.go2proto.test.SumTypeC
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
	(*durationpb.Duration)(nil),   // 8: google.protobuf.Duration
}
var file_model_proto_depIdxs = []int32{
	7,  // 0: xyz.block.ftl.go2proto.test.Message.time:type_name -> google.protobuf.Timestamp
	8,  // 1: xyz.block.ftl.go2proto.test.Message.duration:type_name -> google.protobuf.Duration
	1,  // 2: xyz.block.ftl.go2proto.test.Root.message_ptr:type_name -> xyz.block.ftl.go2proto.test.Message
	0,  // 3: xyz.block.ftl.go2proto.test.Root.enum:type_name -> xyz.block.ftl.go2proto.test.Enum
	3,  // 4: xyz.block.ftl.go2proto.test.Root.sum_type:type_name -> xyz.block.ftl.go2proto.test.SumType
	1,  // 5: xyz.block.ftl.go2proto.test.Root.optional_msg:type_name -> xyz.block.ftl.go2proto.test.Message
	1,  // 6: xyz.block.ftl.go2proto.test.Root.repeated_msg:type_name -> xyz.block.ftl.go2proto.test.Message
	4,  // 7: xyz.block.ftl.go2proto.test.SumType.a:type_name -> xyz.block.ftl.go2proto.test.SumTypeA
	5,  // 8: xyz.block.ftl.go2proto.test.SumType.b:type_name -> xyz.block.ftl.go2proto.test.SumTypeB
	6,  // 9: xyz.block.ftl.go2proto.test.SumType.c:type_name -> xyz.block.ftl.go2proto.test.SumTypeC
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_model_proto_init() }
func file_model_proto_init() {
	if File_model_proto != nil {
		return
	}
	file_model_proto_msgTypes[1].OneofWrappers = []any{}
	file_model_proto_msgTypes[2].OneofWrappers = []any{
		(*SumType_A)(nil),
		(*SumType_B)(nil),
		(*SumType_C)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_model_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_model_proto_goTypes,
		DependencyIndexes: file_model_proto_depIdxs,
		EnumInfos:         file_model_proto_enumTypes,
		MessageInfos:      file_model_proto_msgTypes,
	}.Build()
	File_model_proto = out.File
	file_model_proto_rawDesc = nil
	file_model_proto_goTypes = nil
	file_model_proto_depIdxs = nil
}
