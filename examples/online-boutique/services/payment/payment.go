//ftl:module payment
package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/examples/online-boutique/common/money"
)

type InvalidCreditCardErr struct{}

func (e InvalidCreditCardErr) Error() string { return "invalid credit card" }

type UnacceptedCreditCardErr struct{}

func (e UnacceptedCreditCardErr) Error() string {
	return "credit card not accepted; only VISA or MasterCard are accepted"
}

type ExpiredCreditCardErr struct{}

func (e ExpiredCreditCardErr) Error() string { return "credit card expired" }

type CreditCardInfo struct {
	Number          string
	CVV             int
	ExpirationYear  int
	ExpirationMonth time.Month
}

// LastFour returns the last four digits of the card number.
func (c CreditCardInfo) LastFour() string {
	num := c.Number
	if len(num) > 4 {
		num = num[len(num)-4:]
	}
	return num
}

type ChargeRequest struct {
	Amount     money.Money
	CreditCard CreditCardInfo
}

type ChargeResponse struct {
	TransactionID string
}

//ftl:verb
func Charge(ctx context.Context, req ChargeRequest) (ChargeResponse, error) {
	card := req.CreditCard
	number := strings.ReplaceAll(card.Number, "-", "")
	var company string
	switch {
	case len(number) < 4:
		return ChargeResponse{}, InvalidCreditCardErr{}
	case number[0] == '4':
		company = "Visa"
	case number[0] == '5':
		company = "MasterCard"
	default:
		return ChargeResponse{}, InvalidCreditCardErr{}
	}
	if card.CVV < 100 || card.CVV > 9999 {
		return ChargeResponse{}, InvalidCreditCardErr{}
	}
	if time.Date(card.ExpirationYear, card.ExpirationMonth, 0, 0, 0, 0, 0, time.Local).Before(time.Now()) {
		return ChargeResponse{}, ExpiredCreditCardErr{}
	}

	// Card is valid: process the transaction.
	fmt.Println(
		"Transaction processed",
		"company", company,
		"last_four", number[len(number)-4:],
		"currency", req.Amount.CurrencyCode,
		"amount", fmt.Sprintf("%d.%d", req.Amount.Units, req.Amount.Nanos),
	)
	return ChargeResponse{TransactionID: uuid.New().String()}, nil
}
