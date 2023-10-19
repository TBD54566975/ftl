//ftl:module shipping
package shipping

import (
	"context"
	"fmt"
	"math"

	"github.com/TBD54566975/ftl/examples/online-boutique/common/money"
	"github.com/TBD54566975/ftl/examples/online-boutique/services/cart"
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

//ftl:verb
//ftl:ingress POST /shipping/quote
func GetQuote(ctx context.Context, req ShippingRequest) (money.Money, error) {
	return moneyFromUSD(8.99), nil
}

type ShipOrderResponse struct {
	ID string
}

//ftl:verb
//ftl:ingress POST /shipping/ship
func ShipOrder(ctx context.Context, req ShippingRequest) (ShipOrderResponse, error) {
	baseAddress := fmt.Sprintf("%s, %s, %s", req.Address.StreetAddress, req.Address.City, req.Address.State)
	return ShipOrderResponse{ID: createTrackingID(baseAddress)}, nil
}

func moneyFromUSD(value float64) money.Money {
	units, fraction := math.Modf(value)
	return money.Money{
		CurrencyCode: "USD",
		Units:        int(units),
		Nanos:        int(math.Trunc(fraction * 10000000)),
	}
}
