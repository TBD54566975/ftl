package java

import (
	"github.com/TBD54566975/ftl/jvm-runtime/java"
	"github.com/TBD54566975/ftl/jvm-runtime/plugin/common"
)

func New() *common.Service {
	return common.New(java.Files())
}
