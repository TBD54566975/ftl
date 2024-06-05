package model

import (
	"github.com/alecthomas/types/optional"
)

type Subscription struct {
	Name   string
	Key    SubscriptionKey
	Topic  TopicKey
	Cursor optional.Option[TopicEventKey]
}
