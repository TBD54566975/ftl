package schemapb

// This is a hack around protoc-geng-o oneof's generated code abomination.
// See https://github.com/golang/protobuf/issues/261

type IsDeclValue = isDecl_Value
type IsTypeValue = isType_Value
type IsMetadataValue = isMetadata_Value

func (v *VerbRef) ToFTL() string {
	return v.Module + "." + v.Name
}
