//ftl:module currency
package currency

import (
	"context"
	_ "embed"
	"fmt"
	"math"

	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/examples/online-boutique/common"
	"github.com/TBD54566975/ftl/examples/online-boutique/services/productcatalog"
)

var (
	//go:embed database.json
	databaseJSON []byte
	database     = common.LoadDatabase[map[string]float64](databaseJSON)
)

type GetSupportedCurrenciesRequest struct {
}

type GetSupportedCurrenciesResponse struct {
	CurrencyCodes []string
}

//ftl:verb
func GetSupportedCurrencies(ctx context.Context, req GetSupportedCurrenciesRequest) (GetSupportedCurrenciesResponse, error) {
	return GetSupportedCurrenciesResponse{CurrencyCodes: maps.Keys(database)}, nil
}

type CurrencyConversionRequest struct {
	From   productcatalog.Money
	ToCode string
}

//ftl:verb
func Convert(ctx context.Context, req CurrencyConversionRequest) (productcatalog.Money, error) {
	from := req.From
	fromRate, ok := database[from.CurrencyCode]
	if !ok {
		return productcatalog.Money{}, fmt.Errorf("unknown origin currency %q", req.From.CurrencyCode)
	}
	toRate, ok := database[req.ToCode]
	if !ok {
		return productcatalog.Money{}, fmt.Errorf("unknown destination currency %q", req.ToCode)
	}
	euros := carry(float64(from.Units)/fromRate, float64(from.Nanos)/fromRate)
	to := carry(float64(euros.Units)*toRate, float64(euros.Nanos)*toRate)
	to.CurrencyCode = req.ToCode
	return to, nil
}

// carry is a helper function that handles decimal/fractional carrying.
func carry(units float64, nanos float64) productcatalog.Money {
	const fractionSize = 1000000000 // 1B
	nanos += math.Mod(units, 1.0) * fractionSize
	units = math.Floor(units) + math.Floor(nanos/fractionSize)
	nanos = math.Mod(nanos, fractionSize)
	return productcatalog.Money{
		Units: int(units),
		Nanos: int(nanos),
	}
}
