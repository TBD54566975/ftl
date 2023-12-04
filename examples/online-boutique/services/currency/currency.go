//ftl:module currency
package currency

import (
	"context"
	_ "embed"
	"fmt"
	"math"

	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/examples/online-boutique/common"
)

var (
	//go:embed database.json
	databaseJSON []byte
	database     = common.LoadDatabase[map[string]float64](databaseJSON)
)

// Money represents an amount of money along with the currency type.
type Money struct {
	// The 3-letter currency code defined in ISO 4217.
	CurrencyCode string `json:"currencyCode"`

	// The whole units of the amount.
	// For example if `currencyCode` is `"USD"`, then 1 unit is one US dollar.
	Units int `json:"units"`

	// Number of nano (10^-9) units of the amount.
	// The value must be between -999,999,999 and +999,999,999 inclusive.
	// If `units` is positive, `nanos` must be positive or zero.
	// If `units` is zero, `nanos` can be positive, zero, or negative.
	// If `units` is negative, `nanos` must be negative or zero.
	// For example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	Nanos int `json:"nanos"`
}

type GetSupportedCurrenciesRequest struct {
}

type GetSupportedCurrenciesResponse struct {
	CurrencyCodes []string
}

//ftl:verb
//ftl:ingress GET /currency/supported
func GetSupportedCurrencies(ctx context.Context, req GetSupportedCurrenciesRequest) (GetSupportedCurrenciesResponse, error) {
	return GetSupportedCurrenciesResponse{CurrencyCodes: maps.Keys(database)}, nil
}

type ConvertRequest struct {
	From   Money
	ToCode string
}

//ftl:verb
//ftl:ingress POST /currency/convert
func Convert(ctx context.Context, req ConvertRequest) (Money, error) {
	from := req.From
	fromRate, ok := database[from.CurrencyCode]
	if !ok {
		return Money{}, fmt.Errorf("unknown origin currency %q", req.From.CurrencyCode)
	}
	toRate, ok := database[req.ToCode]
	if !ok {
		return Money{}, fmt.Errorf("unknown destination currency %q", req.ToCode)
	}
	euros := carry(float64(from.Units)/fromRate, float64(from.Nanos)/fromRate)
	to := carry(float64(euros.Units)*toRate, float64(euros.Nanos)*toRate)
	to.CurrencyCode = req.ToCode
	return to, nil
}

// carry is a helper function that handles decimal/fractional carrying.
func carry(units float64, nanos float64) Money {
	const fractionSize = 1000000000 // 1B
	nanos += math.Mod(units, 1.0) * fractionSize
	units = math.Floor(units) + math.Floor(nanos/fractionSize)
	nanos = math.Mod(nanos, fractionSize)
	return Money{
		Units: int(units),
		Nanos: int(nanos),
	}
}
