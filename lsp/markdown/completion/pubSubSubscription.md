Snippet for declaring a subscription to a topic.

```go
var _ = ftl.Subscription(invoicesTopic, "emailInvoices")
```

See https://tbd54566975.github.io/ftl/docs/reference/pubsub/
---
var _ = ftl.Subscription(${1:topicVar}, "${2:subscriptionName}")
