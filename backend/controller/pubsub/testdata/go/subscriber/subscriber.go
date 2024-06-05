package subscriber

import (
	"context"
	"ftl/publisher"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var _ = ftl.Subscription(publisher.Test_topic, "test_subscription")

//ftl:verb
//ftl:subscribe test_subscription
func Echo(ctx context.Context, req publisher.PubSubEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Subscriber is processing %v", req.Time)
	return nil
}
