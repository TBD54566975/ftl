package kotlin

import (
	"github.com/TBD54566975/ftl/jvm-runtime/kotlin"
	"github.com/TBD54566975/ftl/jvm-runtime/plugin/common"
)

func New() *common.Service {
	return common.New(kotlin.Files())
}
