//ftl:module payment
package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"ftl/currency"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type CreditCardInfo struct {
	Number          string
	CVV             int
	ExpirationYear  int
	ExpirationMonth int
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
	Amount     currency.Money
	CreditCard CreditCardInfo
}

type ChargeResponse struct {
	TransactionID string
}

//ftl:ingress POST /payment/charge
func Charge(ctx context.Context, req builtin.HttpRequest[ChargeRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[ChargeResponse, ErrorResponse], error) {
	card := req.Body.CreditCard
	number := strings.ReplaceAll(card.Number, "-", "")
	var company string
	switch {
	case len(number) < 4:
		return builtin.HttpResponse[ChargeResponse, ErrorResponse]{Error: ftl.Some(ErrorResponse{Message: "Invalid card number"})}, nil
	case number[0] == '4':
		company = "Visa"
	case number[0] == '5':
		company = "MasterCard"
	default:
		return builtin.HttpResponse[ChargeResponse, ErrorResponse]{Error: ftl.Some(ErrorResponse{Message: "Invalid card number"})}, nil
	}
	if card.CVV < 100 || card.CVV > 9999 {
		return builtin.HttpResponse[ChargeResponse, ErrorResponse]{Error: ftl.Some(ErrorResponse{Message: "Invalid CVV number"})}, nil
	}
	if time.Date(card.ExpirationYear, time.Month(card.ExpirationMonth), 0, 0, 0, 0, 0, time.Local).Before(time.Now()) {
		return builtin.HttpResponse[ChargeResponse, ErrorResponse]{Error: ftl.Some(ErrorResponse{Message: "Card expired"})}, nil
	}

	// Card is valid: process the transaction.
	fmt.Println(
		"Transaction processed",
		"company", company,
		"last_four", number[len(number)-4:],
		"currency", req.Body.Amount.CurrencyCode,
		"amount", fmt.Sprintf("%d.%d", req.Body.Amount.Units, req.Body.Amount.Nanos),
	)
	return builtin.HttpResponse[ChargeResponse, ErrorResponse]{
		Body: ftl.Some(ChargeResponse{TransactionID: uuid.New().String()}),
	}, nil
}
