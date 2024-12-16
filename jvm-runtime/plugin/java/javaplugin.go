package java

import (
	"github.com/block/ftl/jvm-runtime/java"
	"github.com/block/ftl/jvm-runtime/plugin/common"
)

func New() *common.Service {
	return common.New(java.Files())
}
