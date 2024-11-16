package schemaservice

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/schema"
)

func TestStateValidation(t *testing.T) {
	t.Parallel()

	state := NewState()

	validModule := &schema.Module{
		Name: "time",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "time",
				Export:   true,
				Request:  &schema.Unit{},
				Response: &schema.Time{},
			},
		},
	}

	err := state.UpsertModule(validModule)
	assert.NoError(t, err)

	invalidModule := &schema.Module{
		Name: "echo",
		Decls: []schema.Decl{
			&schema.Verb{
				Name:     "echo",
				Request:  &schema.Unit{},
				Response: &schema.String{},
				Metadata: []schema.Metadata{
					&schema.MetadataCalls{
						Calls: []*schema.Ref{{Module: "time", Name: "invalid"}},
					},
				},
			},
		},
	}

	err = state.UpsertModule(invalidModule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `unknown reference "time.invalid"`)

	invalidModule.Decls[0].(*schema.Verb).Metadata[0].(*schema.MetadataCalls).Calls[0].Name = "time" //nolint:forcetypeassert

	err = state.UpsertModule(invalidModule)
	assert.NoError(t, err)

	_, err = state.DeleteModule("time")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `unknown reference "time.time"`)
}
