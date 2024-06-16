Snippet for a sink function that consume events from a PubSub subscription.

```go
//ftl:subscribe emailInvoices
func SendInvoiceEmail(ctx context.Context, in Invoice) error {
  // ...
}
```

See https://tbd54566975.github.io/ftl/docs/reference/pubsub/
---
//ftl:subscribe ${1:subscriptionName}
func ${2:FunctionName}(ctx context.Context, in ${3:Type}) error {
	${4:// TODO: Implement}
	return nil
}
