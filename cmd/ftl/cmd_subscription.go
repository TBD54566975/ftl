package main

type subscriptionCmd struct {
	Reset resetSubscriptionCmd `cmd:"" help:"Reset the subscription to the head of its topic."`
}
