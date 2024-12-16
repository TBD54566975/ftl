package child

import (
	"github.com/block/ftl/go-runtime/ftl"
	lib "github.com/block/ftl/go-runtime/schema/testdata"
)

type BadChildStruct struct {
	Body uint64
}

//ftl:data
type UnaliasedExternalType struct {
	Field lib.NonFTLType
}

//ftl:typealias
//ftl:typemap go "github.com/blah.lib.NonFTLType"
type WrongMappingExternal lib.NonFTLType

//ftl:typealias
//ftl:typemap go "github.com/block/ftl/go-runtime/schema/testdata.lib.NonFTLType"
//ftl:typemap go "github.com/block/ftl/go-runtime/schema/testdata.lib.NonFTLType"
type MultipleMappings lib.NonFTLType

//ftl:data
type Redeclared struct {
}

//ftl:enum
type EnumVariantConflictChild int

const (
	SameVariant EnumVariantConflictChild = iota
)

type FtlConfigEndpoint = ftl.Config[string]
type FtlSecretEndpoint = ftl.Secret[string]

type DifferentDeclDupl = ftl.Config[string]

type DuplDbConfig struct {
	ftl.DefaultPostgresDatabaseConfig
}

func (DuplDbConfig) Name() string { return "testdb" }
