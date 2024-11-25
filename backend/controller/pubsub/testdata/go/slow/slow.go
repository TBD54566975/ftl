package slow

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Topic = ftl.TopicHandle[Event]
type SlowSubscription = ftl.SubscriptionHandle[Topic, ConsumeClient, Event]

type Event struct {
	Duration int
}

type PublishRequest struct {
	Durations []int
}

//ftl:verb
func Publish(ctx context.Context, req PublishRequest, topic Topic) error {
	for _, duration := range req.Durations {
		err := topic.Publish(ctx, Event{Duration: duration})
		if err != nil {
			return err
		}
	}
	return nil
}

//ftl:verb
func Consume(ctx context.Context, event Event) error {
	for i := range event.Duration {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			appendLog("slept for %ds", i+1)
		}
	}
	return nil
}

func appendLog(msg string, args ...interface{}) {
	dest, ok := os.LookupEnv("TEST_LOG_FILE")
	if !ok {
		panic("TEST_LOG_FILE not set")
	}
	w, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, msg+"\n", args...)
	err = w.Close()
	if err != nil {
		panic(err)
	}
}
