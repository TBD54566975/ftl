package encryption

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

// Basic call
//
// Used to test encryption of call event logs

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}

//ftl:data
type Event struct {
	Name string `json:"name"`
}

// PubSub
//
// Used to test encryption of topic_events and async_calls tables

var Topic = ftl.Topic[Event]("topic")
var _ = ftl.Subscription(Topic, "subscription")

//ftl:verb
func Publish(ctx context.Context, e Event) error {
	fmt.Printf("Publishing event: %s\n", e.Name)
	return Topic.Publish(ctx, e)
}

//ftl:verb
//ftl:subscribe subscription
func Consume(ctx context.Context, e Event) error {
	fmt.Printf("Received event: %s\n", e.Name)
	if e.Name != "AliceInWonderland" {
		return fmt.Errorf("Unexpected event: %s", e.Name)
	}
	return nil
}

// FSM
//
// Used to test encryption of async_calls tables via FSM operations

var fsm = ftl.FSM("payment",
	ftl.Start(Created),
	ftl.Start(Paid),
	ftl.Transition(Created, Paid),
	ftl.Transition(Paid, Completed),
)

type OnlinePaymentCompleted struct {
	Name string `json:"name"`
}
type OnlinePaymentPaid struct {
	Name string `json:"name"`
}
type OnlinePaymentCreated struct {
	Name string `json:"name"`
}

//ftl:verb
func BeginFSM(ctx context.Context, req OnlinePaymentCreated) error {
	return fsm.Send(ctx, "test", req)
}

//ftl:verb
func TransitionFSM(ctx context.Context, req OnlinePaymentPaid) error {
	return fsm.Send(ctx, "test", req)
}

//ftl:verb
func Completed(ctx context.Context, in OnlinePaymentCompleted) error {
	return nil
}

//ftl:verb
func Created(ctx context.Context, in OnlinePaymentCreated) error {
	return nil
}

//ftl:verb
func Paid(ctx context.Context, in OnlinePaymentPaid) error {
	return nil
}
