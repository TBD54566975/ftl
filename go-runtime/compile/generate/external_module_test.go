package generate

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestCodegen(t *testing.T) {
	w := &strings.Builder{}
	module := `
		module basket {
			data ItemRequest {
				basketID String
				itemID String
			}
			data BasketSummary {
				items Int
			}
			// Add an item to the basket.
			verb Add(ItemRequest) BasketSummary
			// Remove an item from the basket.
			verb Remove(ItemRequest) BasketSummary
		}`
	s, err := schema.ParseString("", module)
	assert.NoError(t, err)
	err = ExternalModule(w, s.Modules[0])
	assert.NoError(t, err)
	expected := `//ftl:module basket
package basket

import (
  "context"
)

type ItemRequest struct {
  BasketID string ` + "`json:\"basketID\"`" + `
  ItemID string ` + "`json:\"itemID\"`" + `
}

type BasketSummary struct {
  Items int ` + "`json:\"items\"`" + `
}

// Add an item to the basket.
//
//ftl:verb
func Add(context.Context, ItemRequest) (BasketSummary, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}

// Remove an item from the basket.
//
//ftl:verb
func Remove(context.Context, ItemRequest) (BasketSummary, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}
`
	assert.Equal(t, expected, w.String())
}
