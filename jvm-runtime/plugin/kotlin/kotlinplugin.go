package kotlin

import (
	"github.com/block/ftl/jvm-runtime/kotlin"
	"github.com/block/ftl/jvm-runtime/plugin/common"
)

func New() *common.Service {
	return common.New(kotlin.Files())
}
