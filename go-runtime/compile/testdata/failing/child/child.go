package child

import lib "github.com/TBD54566975/ftl/go-runtime/compile/testdata"

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
//ftl:typemap go "github.com/TBD54566975/ftl/go-runtime/compile/testdata.lib.NonFTLType"
//ftl:typemap go "github.com/TBD54566975/ftl/go-runtime/compile/testdata.lib.NonFTLType"
type MultipleMappings lib.NonFTLType
