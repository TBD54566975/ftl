package modulecontext

import (
	"context"
	"testing"

	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

type jsonableStruct struct {
	Str        string          `json:"string"`
	TestStruct *jsonableStruct `json:"struct,omitempty"`
}

func TestConfigManagerInAndOut(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm, err := newInMemoryConfigManager[cf.Configuration](ctx)
	assert.NoError(t, err)

	moduleName := "test"
	strValue := "HelloWorld"
	intValue := 42
	structValue := jsonableStruct{Str: "HelloWorld", TestStruct: &jsonableStruct{Str: "HelloWorld"}}
	cm.Set(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "str"}, strValue)
	cm.Set(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "int"}, intValue)
	cm.Set(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "struct"}, structValue)

	builder := NewBuilder(moduleName).AddConfigFromManager(cm)

	moduleCtx, err := builder.Build(ctx)
	assert.NoError(t, err)

	var outStr string
	moduleCtx.configManager.Get(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "str"}, &outStr)
	assert.Equal(t, strValue, outStr, "expected string value to be set and retrieved correctly")

	var outInt int
	moduleCtx.configManager.Get(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "int"}, &outInt)
	assert.Equal(t, intValue, outInt, "expected int value to be set and retrieved correctly")

	var outStruct jsonableStruct
	moduleCtx.configManager.Get(ctx, cf.Ref{Module: optional.Some(moduleName), Name: "struct"}, &outStruct)
	assert.Equal(t, structValue, outStruct, "expected struct value to be set and retrieved correctly")
}
