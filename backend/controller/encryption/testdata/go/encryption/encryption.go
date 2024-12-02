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

type Topic = ftl.TopicHandle[Event, ftl.SinglePartitionMap[Event]]

//ftl:verb
func Publish(ctx context.Context, e Event, topic Topic) error {
	fmt.Printf("Publishing event: %s\n", e.Name)
	return topic.Publish(ctx, e)
}

//ftl:verb
//ftl:subscribe topic from=beginning
func Consume(ctx context.Context, e Event) error {
	fmt.Printf("Received event: %s\n", e.Name)
	if e.Name != "AliceInWonderland" {
		return fmt.Errorf("Unexpected event: %s", e.Name)
	}
	return nil
}
