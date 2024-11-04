package child

import (
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	lib "github.com/TBD54566975/ftl/go-runtime/schema/testdata"
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
//ftl:typemap go "github.com/TBD54566975/ftl/go-runtime/schema/testdata.lib.NonFTLType"
//ftl:typemap go "github.com/TBD54566975/ftl/go-runtime/schema/testdata.lib.NonFTLType"
type MultipleMappings lib.NonFTLType

//ftl:data
type Redeclared struct {
}

//ftl:enum
type EnumVariantConflictChild int

const (
	SameVariant EnumVariantConflictChild = iota
)

var duplConfig = ftl.Config[string]("FTL_CONFIG_ENDPOINT")
var duplSecret = ftl.Secret[string]("FTL_SECRET_ENDPOINT")

var duplicateDeclName = ftl.Config[string]("PrivateData")

type DuplDbConfig struct {
	ftl.DefaultPostgresDatabaseConfig
}

func (DuplDbConfig) Name() string { return "testdb" }
