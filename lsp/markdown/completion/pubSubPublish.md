Snippet for publishing an event to a topic.

```go
invoicesTopic.Publish(ctx, Invoice{...})
```

See https://tbd54566975.github.io/ftl/docs/reference/pubsub/
---
${1:topicVar}.Publish(ctx, ${2:Type}{${3:...}})
```