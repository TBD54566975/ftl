package subscriber

import (
	"ftl/pubsub"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var _ = ftl.Subscription(pubsub.Topic1, "subscription1_1")
