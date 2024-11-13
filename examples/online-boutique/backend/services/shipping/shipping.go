//ftl:module shipping
package shipping

import (
	"context"
	"fmt"
	"math"

	"ftl/currency"

	"ftl/builtin"
	"ftl/cart"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type Address struct {
	StreetAddress string
	City          string
	State         string
	Country       string
	ZipCode       int
}

type ShippingRequest struct {
	Address Address
	Items   []cart.Item
}

//ftl:ingress POST /shipping/quote
func GetQuote(ctx context.Context, req builtin.HttpRequest[ShippingRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[currency.Money, ftl.Unit], error) {
	return builtin.HttpResponse[currency.Money, ftl.Unit]{Body: ftl.Some(moneyFromUSD(8.99))}, nil
}

type ShipOrderResponse struct {
	ID string
}

//ftl:ingress POST /shipping/ship
func ShipOrder(ctx context.Context, req builtin.HttpRequest[ShippingRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[ShipOrderResponse, ftl.Unit], error) {
	baseAddress := fmt.Sprintf("%s, %s, %s", req.Body.Address.StreetAddress, req.Body.Address.City, req.Body.Address.State)
	return builtin.HttpResponse[ShipOrderResponse, ftl.Unit]{Body: ftl.Some(ShipOrderResponse{ID: createTrackingID(baseAddress)})}, nil
}

func moneyFromUSD(value float64) currency.Money {
	units, fraction := math.Modf(value)
	return currency.Money{
		CurrencyCode: "USD",
		Units:        int(units),
		Nanos:        int(math.Trunc(fraction * 10000000)),
	}
}
