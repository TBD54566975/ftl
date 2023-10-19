//ftl:module currency
package currency

import (
	"context"
	_ "embed"
	"fmt"
	"math"

	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/examples/online-boutique/common"
	"github.com/TBD54566975/ftl/examples/online-boutique/common/money"
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
//ftl:ingress GET /currency/supported
func GetSupportedCurrencies(ctx context.Context, req GetSupportedCurrenciesRequest) (GetSupportedCurrenciesResponse, error) {
	return GetSupportedCurrenciesResponse{CurrencyCodes: maps.Keys(database)}, nil
}

type ConvertRequest struct {
	From   money.Money
	ToCode string
}

//ftl:verb
//ftl:ingress POST /currency/convert
func Convert(ctx context.Context, req ConvertRequest) (money.Money, error) {
	from := req.From
	fromRate, ok := database[from.CurrencyCode]
	if !ok {
		return money.Money{}, fmt.Errorf("unknown origin currency %q", req.From.CurrencyCode)
	}
	toRate, ok := database[req.ToCode]
	if !ok {
		return money.Money{}, fmt.Errorf("unknown destination currency %q", req.ToCode)
	}
	euros := carry(float64(from.Units)/fromRate, float64(from.Nanos)/fromRate)
	to := carry(float64(euros.Units)*toRate, float64(euros.Nanos)*toRate)
	to.CurrencyCode = req.ToCode
	return to, nil
}

// carry is a helper function that handles decimal/fractional carrying.
func carry(units float64, nanos float64) money.Money {
	const fractionSize = 1000000000 // 1B
	nanos += math.Mod(units, 1.0) * fractionSize
	units = math.Floor(units) + math.Floor(nanos/fractionSize)
	nanos = math.Mod(nanos, fractionSize)
	return money.Money{
		Units: int(units),
		Nanos: int(nanos),
	}
}
