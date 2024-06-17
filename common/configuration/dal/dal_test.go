package dal

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestModuleConfiguration(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)
	assert.NotZero(t, dal)

	tests := []struct {
		TestName     string
		ModuleSet    optional.Option[string]
		ModuleGet    optional.Option[string]
		PresetGlobal bool
	}{
		{
			"SetModuleGetModule",
			optional.Some("echo"),
			optional.Some("echo"),
			false,
		},
		{
			"SetGlobalGetGlobal",
			optional.None[string](),
			optional.None[string](),
			false,
		},
		{
			"SetGlobalGetModule",
			optional.None[string](),
			optional.Some("echo"),
			false,
		},
		{
			"SetModuleOverridesGlobal",
			optional.Some("echo"),
			optional.Some("echo"),
			true,
		},
	}

	b := []byte(`"asdf"`)
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			if test.PresetGlobal {
				err := dal.SetModuleConfiguration(ctx, optional.None[string](), "configname", []byte(`"qwerty"`))
				assert.NoError(t, err)
			}
			err := dal.SetModuleConfiguration(ctx, test.ModuleSet, "configname", b)
			assert.NoError(t, err)
			gotBytes, err := dal.GetModuleConfiguration(ctx, test.ModuleGet, "configname")
			assert.NoError(t, err)
			assert.Equal(t, b, gotBytes)
			err = dal.UnsetModuleConfiguration(ctx, test.ModuleGet, "configname")
			assert.NoError(t, err)
		})
	}

	t.Run("List", func(t *testing.T) {
		sortedList := []sql.ModuleConfiguration{
			{
				Module: optional.Some("echo"),
				Name:   "a",
			},
			{
				Module: optional.Some("echo"),
				Name:   "b",
			},
			{
				Module: optional.None[string](),
				Name:   "a",
			},
		}

		// Insert entries in a separate order from how they should be returned to
		// test sorting logic in the SQL query
		err := dal.SetModuleConfiguration(ctx, sortedList[1].Module, sortedList[1].Name, []byte(`""`))
		assert.NoError(t, err)
		err = dal.SetModuleConfiguration(ctx, sortedList[2].Module, sortedList[2].Name, []byte(`""`))
		assert.NoError(t, err)
		err = dal.SetModuleConfiguration(ctx, sortedList[0].Module, sortedList[0].Name, []byte(`""`))
		assert.NoError(t, err)

		gotList, err := dal.ListModuleConfiguration(ctx)
		assert.NoError(t, err)
		for i := range sortedList {
			assert.Equal(t, sortedList[i].Module, gotList[i].Module)
			assert.Equal(t, sortedList[i].Name, gotList[i].Name)
		}
	})
}
