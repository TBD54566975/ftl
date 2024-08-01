package main

type pubsubCmd struct {
	Subscription subscriptionCmd `cmd:"" help:"Manage subscriptions."`
}
