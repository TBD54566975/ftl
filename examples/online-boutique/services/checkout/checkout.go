//ftl:module checkout
package checkout

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"ftl/builtin"
	"ftl/cart"
	"ftl/currency"
	"ftl/payment"
	"ftl/productcatalog"
	"ftl/shipping"

	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/examples/online-boutique/common/money"

	ftl "github.com/TBD54566975/ftl/go-runtime/sdk"
)

type PlaceOrderRequest struct {
	UserID       string
	UserCurrency string

	Address    shipping.Address
	Email      string
	CreditCard payment.CreditCardInfo
}

type OrderItem struct {
	Item cart.Item
	Cost currency.Money
}

type Order struct {
	ID                 string
	ShippingTrackingID string
	ShippingCost       currency.Money
	ShippingAddress    shipping.Address
	Items              []OrderItem
}

//ftl:verb
//ftl:ingress POST /checkout/{userID}
func PlaceOrder(ctx context.Context, req builtin.HttpRequest[PlaceOrderRequest]) (builtin.HttpResponse[Order], error) {
	cartItems, err := ftl.Call(ctx, cart.GetCart, builtin.HttpRequest[cart.GetCartRequest]{Body: cart.GetCartRequest{UserID: req.Body.UserID}})
	if err != nil {
		return builtin.HttpResponse[Order]{}, fmt.Errorf("failed to get cart: %w", err)
	}

	orders := make([]OrderItem, len(cartItems.Body.Items))
	for i, item := range cartItems.Body.Items {
		products, err := ftl.Call(ctx, productcatalog.Get, builtin.HttpRequest[productcatalog.GetRequest]{Body: productcatalog.GetRequest{Id: item.ProductID}})
		if err != nil {
			return builtin.HttpResponse[Order]{}, fmt.Errorf("failed to get product #%q: %w", item.ProductID, err)
		}
		price, err := ftl.Call(ctx, currency.Convert, builtin.HttpRequest[currency.ConvertRequest]{
			Body: currency.ConvertRequest{
				From:   products.Body.PriceUSD,
				ToCode: req.Body.UserCurrency,
			},
		})
		if err != nil {
			return builtin.HttpResponse[Order]{}, fmt.Errorf("failed to convert price of %q to %s: %w", item.ProductID, req.Body.UserCurrency, err)
		}
		orders[i] = OrderItem{
			Item: item,
			Cost: price.Body,
		}
	}

	shippingUSD, err := ftl.Call(ctx, shipping.GetQuote, builtin.HttpRequest[shipping.ShippingRequest]{
		Body: shipping.ShippingRequest{
			Address: req.Body.Address,
			Items:   slices.Map(orders, func(i OrderItem) cart.Item { return i.Item }),
		},
	})
	if err != nil {
		return builtin.HttpResponse[Order]{}, fmt.Errorf("failed to get shipping quote: %w", err)
	}
	shippingPrice, err := ftl.Call(ctx, currency.Convert, builtin.HttpRequest[currency.ConvertRequest]{
		Body: currency.ConvertRequest{
			From:   shippingUSD.Body,
			ToCode: req.Body.UserCurrency,
		},
	})
	if err != nil {
		return builtin.HttpResponse[Order]{}, fmt.Errorf("failed to convert shipping cost to currency: %w", err)
	}

	total := currency.Money{CurrencyCode: req.Body.UserCurrency}
	total = money.Must(money.Sum(total, shippingPrice.Body))
	for _, it := range orders {
		multPrice := money.MultiplySlow(it.Cost, uint32(it.Item.Quantity))
		total = money.Must(money.Sum(total, multPrice))
	}
	txID, err := ftl.Call(ctx, payment.Charge, builtin.HttpRequest[payment.ChargeRequest]{
		Body: payment.ChargeRequest{
			Amount:     total,
			CreditCard: req.Body.CreditCard,
		},
	})
	if err != nil {
		return builtin.HttpResponse[Order]{}, fmt.Errorf("failed to charge card: %w", err)
	}
	fmt.Printf("Charged card, ID %s", txID.Body.TransactionID)

	shippingTrackingID, err := ftl.Call(ctx, shipping.ShipOrder, builtin.HttpRequest[shipping.ShippingRequest]{
		Body: shipping.ShippingRequest{
			Address: req.Body.Address,
			Items:   cartItems.Body.Items,
		},
	})
	if err != nil {
		return builtin.HttpResponse[Order]{}, fmt.Errorf("shipping error: %w", err)
	}
	fmt.Printf("Shipped order, tracking ID %s", shippingTrackingID.Body.Id)

	// Empty the cart, but don't worry about errors.
	_, _ = ftl.Call(ctx, cart.EmptyCart, builtin.HttpRequest[cart.EmptyCartRequest]{Body: cart.EmptyCartRequest{UserID: req.Body.UserID}})

	order := Order{
		ID:                 uuid.New().String(),
		ShippingTrackingID: shippingTrackingID.Body.Id,
		ShippingCost:       shippingPrice.Body,
		ShippingAddress:    req.Body.Address,
		Items:              orders,
	}
	// if err := s.emailService.Get().SendOrderConfirmation(ctx, req.Email, order); err != nil {
	// 	s.Logger(ctx).Error("failed to send order confirmation", "err", err, "email", req.Email)
	// } else {
	// 	s.Logger(ctx).Info("order confirmation email sent", "email", req.Email)
	// }

	return builtin.HttpResponse[Order]{Body: order}, nil
}
