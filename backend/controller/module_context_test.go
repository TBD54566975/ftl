package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/TBD54566975/ftl/backend/schema"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestModuleContextProto(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	moduleName := "test"

	cp := cf.NewInMemoryProvider[cf.Configuration]()
	cr := cf.NewInMemoryResolver[cf.Configuration]()
	cm, err := cf.New(ctx, cr, []cf.Provider[cf.Configuration]{cp})
	assert.NoError(t, err)
	ctx = cf.ContextWithConfig(ctx, cm)

	sp := cf.NewInMemoryProvider[cf.Secrets]()
	sr := cf.NewInMemoryResolver[cf.Secrets]()
	sm, err := cf.New(ctx, sr, []cf.Provider[cf.Secrets]{sp})
	assert.NoError(t, err)
	ctx = cf.ContextWithSecrets(ctx, sm)

	// Set 50 configs and 50 global configs
	// It's hard to tell if module config beats global configs because we are dealing with unordered maps, or because the logic is correct
	// Repeating it 50 times hopefully gives us a good chance of catching inconsistencies
	for i := range 50 {
		key := fmt.Sprintf("key%d", i)

		strValue := "HelloWorld"
		globalStrValue := "GlobalHelloWorld"
		assert.NoError(t, cm.Set(ctx, cf.Ref{Module: optional.Some(moduleName), Name: key}, strValue))
		assert.NoError(t, cm.Set(ctx, cf.Ref{Module: optional.None[string](), Name: key}, globalStrValue))
	}

	response, err := moduleContextToProto(ctx, moduleName, []*schema.Module{
		{
			Name: moduleName,
		},
	})
	assert.NoError(t, err)

	for i := range 50 {
		key := fmt.Sprintf("key%d", i)
		assert.Equal(t, "\"HelloWorld\"", string(response.Msg.Configs[key]), "module configs should beat global configs")
	}
}
