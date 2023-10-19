//ftl:module checkout
package checkout

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/examples/online-boutique/common/money"
	"github.com/TBD54566975/ftl/examples/online-boutique/services/cart"
	"github.com/TBD54566975/ftl/examples/online-boutique/services/currency"
	"github.com/TBD54566975/ftl/examples/online-boutique/services/payment"
	"github.com/TBD54566975/ftl/examples/online-boutique/services/productcatalog"
	"github.com/TBD54566975/ftl/examples/online-boutique/services/shipping"
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
	Cost money.Money
}

type Order struct {
	ID                 string
	ShippingTrackingID string
	ShippingCost       money.Money
	ShippingAddress    shipping.Address
	Items              []OrderItem
}

//ftl:verb
//ftl:ingress POST /checkout
func PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error) {
	cartItems, err := ftl.Call(ctx, cart.GetCart, cart.GetCartRequest{UserID: req.UserID})
	if err != nil {
		return Order{}, fmt.Errorf("failed to get cart: %w", err)
	}

	orders := make([]OrderItem, len(cartItems.Items))
	for i, item := range cartItems.Items {
		product, err := ftl.Call(ctx, productcatalog.Get, productcatalog.GetRequest{ID: item.ProductID})
		if err != nil {
			return Order{}, fmt.Errorf("failed to get product #%q: %w", item.ProductID, err)
		}
		price, err := ftl.Call(ctx, currency.Convert, currency.ConvertRequest{
			From:   product.PriceUSD,
			ToCode: req.UserCurrency,
		})
		if err != nil {
			return Order{}, fmt.Errorf("failed to convert price of %q to %s: %w", item.ProductID, req.UserCurrency, err)
		}
		orders[i] = OrderItem{
			Item: item,
			Cost: price,
		}
	}

	shippingUSD, err := ftl.Call(ctx, shipping.GetQuote, shipping.ShippingRequest{
		Address: req.Address,
		Items:   slices.Map(orders, func(i OrderItem) cart.Item { return i.Item }),
	})
	if err != nil {
		return Order{}, fmt.Errorf("failed to get shipping quote: %w", err)
	}
	shippingPrice, err := ftl.Call(ctx, currency.Convert, currency.ConvertRequest{
		From:   shippingUSD,
		ToCode: req.UserCurrency,
	})
	if err != nil {
		return Order{}, fmt.Errorf("failed to convert shipping cost to currency: %w", err)
	}

	total := money.Money{CurrencyCode: req.UserCurrency}
	total = money.Must(money.Sum(total, shippingPrice))
	for _, it := range orders {
		multPrice := money.MultiplySlow(it.Cost, uint32(it.Item.Quantity))
		total = money.Must(money.Sum(total, multPrice))
	}
	txID, err := ftl.Call(ctx, payment.Charge, payment.ChargeRequest{
		Amount:     total,
		CreditCard: req.CreditCard,
	})
	if err != nil {
		return Order{}, fmt.Errorf("failed to charge card: %w", err)
	}
	fmt.Printf("Charged card, ID %s", txID.TransactionID)

	shippingTrackingID, err := ftl.Call(ctx, shipping.ShipOrder, shipping.ShippingRequest{
		Address: req.Address,
		Items:   cartItems.Items,
	})
	if err != nil {
		return Order{}, fmt.Errorf("shipping error: %w", err)
	}
	fmt.Printf("Shipped order, tracking ID %s", shippingTrackingID.ID)

	// Empty the cart, but don't worry about errors.
	_, _ = ftl.Call(ctx, cart.EmptyCart, cart.EmptyCartRequest{UserID: req.UserID})

	order := Order{
		ID:                 uuid.New().String(),
		ShippingTrackingID: shippingTrackingID.ID,
		ShippingCost:       shippingPrice,
		ShippingAddress:    req.Address,
		Items:              orders,
	}
	// if err := s.emailService.Get().SendOrderConfirmation(ctx, req.Email, order); err != nil {
	// 	s.Logger(ctx).Error("failed to send order confirmation", "err", err, "email", req.Email)
	// } else {
	// 	s.Logger(ctx).Info("order confirmation email sent", "email", req.Email)
	// }

	return order, nil
}
