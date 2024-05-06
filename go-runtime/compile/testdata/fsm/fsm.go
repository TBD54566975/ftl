package fsm

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var paymentFSM = ftl.FSM("payment",
	ftl.Start(Created),
	ftl.Start(Paid),
	ftl.Transition(Created, Paid),
	ftl.Transition(Created, Failed),
	ftl.Transition(Paid, Completed),
)

type OnlinePaymentCompleted struct{}
type OnlinePaymentFailed struct{}
type OnlinePaymentPaid struct{}
type OnlinePaymentCreated struct{}

//ftl:verb
func Completed(ctx context.Context, in OnlinePaymentCompleted) error {
	return nil
}

//ftl:verb
func Created(ctx context.Context, in OnlinePaymentCreated) error {
	return nil
}

//ftl:verb
func Failed(ctx context.Context, in OnlinePaymentFailed) error {
	return nil
}

//ftl:verb
func Paid(ctx context.Context, in OnlinePaymentPaid) error {
	return nil
}
