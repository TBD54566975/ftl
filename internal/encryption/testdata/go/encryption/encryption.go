package encryption

import (
	"context"
	"fmt"
	"time"

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
	ftl.Transition(Created, NextAndSleep),
	ftl.Transition(NextAndSleep, Completed),
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

type NextAndSleepEvent struct {
	Name string `json:"name"`
}

//ftl:verb
func BeginFSM(ctx context.Context, req OnlinePaymentCreated) error {
	return fsm.Send(ctx, req.Name, req)
}

//ftl:verb
func TransitionToPaid(ctx context.Context, req OnlinePaymentPaid) error {
	return fsm.Send(ctx, req.Name, req)
}

//ftl:verb
func TransitionToNextAndSleep(ctx context.Context, req NextAndSleepEvent) error {
	return fsm.Send(ctx, req.Name, req)
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

// NextAndSleep calls fsm.Next() and then sleeps so we can test what is put into the fsm next event table
//
//ftl:verb
func NextAndSleep(ctx context.Context, in NextAndSleepEvent) error {
	err := ftl.FSMNext(ctx, OnlinePaymentCompleted{Name: in.Name})
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Minute)
	return nil
}
