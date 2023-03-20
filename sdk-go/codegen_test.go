package sdkgo

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/schema"
)

func TestCodegen(t *testing.T) {
	w := &strings.Builder{}
	module := `
		module basket {
			data ItemRequest {
				basketID string
				itemID string
			}
			data BasketSummary {
				items int
			}
			// Add an item to the basket.
			verb Add(ItemRequest) BasketSummary
			// Remove an item from the basket.
			verb Remove(ItemRequest) BasketSummary
		}`
	s, err := schema.ParseString("", module)
	assert.NoError(t, err)
	err = Generate(s.Modules[0], w)
	assert.NoError(t, err)
	expected := `package basket

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
func Add(ctx context.Context, ItemRequest) (BasketSummary, error) {
  panic("Verb stubs should not be called directly, instead use sdkgo.Call()")
}

// Remove an item from the basket.
func Remove(ctx context.Context, ItemRequest) (BasketSummary, error) {
  panic("Verb stubs should not be called directly, instead use sdkgo.Call()")
}
`
	assert.Equal(t, expected, w.String())
}
