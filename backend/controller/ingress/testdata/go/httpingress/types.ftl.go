// Code generated by FTL. DO NOT EDIT.
package httpingress

import (
    lib "github.com/TBD54566975/ftl/go-runtime/schema/testdata"

    "github.com/TBD54566975/ftl/go-runtime/ftl/reflection"


)

func init() {
	reflection.Register(
		reflection.SumType[SumType](
			*new(A),
			*new(B),
		),
		reflection.ExternalType(*new(lib.NonFTLType)),
	)
}
