// Code generated by go2proto. DO NOT EDIT.

package testdata

import "fmt"
import destpb "github.com/TBD54566975/ftl/cmd/go2proto/testdata/testdatapb"
import "google.golang.org/protobuf/proto"
import "google.golang.org/protobuf/types/known/timestamppb"
import "google.golang.org/protobuf/types/known/durationpb"

var _ fmt.Stringer
var _ = timestamppb.Timestamp{}
var _ = durationpb.Duration{}


func (x Enum) ToProto() destpb.Enum {
	return destpb.Enum(x)
}

func (x *Message) ToProto() *destpb.Message {
	out := &destpb.Message{}
	out.Time = timestamppb.New(x.Time)
	out.Duration = durationpb.New(x.Duration)
	return out
}

func (x *Root) ToProto() *destpb.Root {
	out := &destpb.Root{}
	out.Int = int64(x.Int)
	out.String_ = string(x.String)
	out.MessagePtr = x.MessagePtr.ToProto()
	out.Enum = x.Enum.ToProto()
	out.SumType = SumTypeToProto(x.SumType)
	out.OptionalInt = proto.Int64(int64(x.OptionalInt))
	out.OptionalIntPtr = proto.Int64(int64(*x.OptionalIntPtr))
	out.OptionalMsg = x.OptionalMsg.ToProto()
	return out
}

func SumTypeToProto(value SumType) *destpb.SumType {
	switch value := value.(type) {
	case *SumTypeA:
		return &destpb.SumType{
			Value: &destpb.SumType_A{value.ToProto()},
		}
	case *SumTypeB:
		return &destpb.SumType{
			Value: &destpb.SumType_B{value.ToProto()},
		}
	case *SumTypeC:
		return &destpb.SumType{
			Value: &destpb.SumType_C{value.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func (x *SumTypeA) ToProto() *destpb.SumTypeA {
	out := &destpb.SumTypeA{}
	out.A = string(x.A)
	return out
}

func (x *SumTypeB) ToProto() *destpb.SumTypeB {
	out := &destpb.SumTypeB{}
	out.B = int64(x.B)
	return out
}

func (x *SumTypeC) ToProto() *destpb.SumTypeC {
	out := &destpb.SumTypeC{}
	out.C = float64(x.C)
	return out
}

		