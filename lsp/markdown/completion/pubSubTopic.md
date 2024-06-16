Snippet for declaring a topic.

```go
var invoicesTopic = ftl.Topic[Invoice]("invoices")
```

See https://tbd54566975.github.io/ftl/docs/reference/pubsub/
---
var ${1:topicVar} = ftl.Topic[${2:Type}]("${1:topicName}")
