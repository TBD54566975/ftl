Declare a PubSub topic.

```go
var Invoices = ftl.Topic[Invoice]("invoices")
```

See https://block.github.io/ftl/docs/reference/pubsub/
---

var ${1:topicVar} = ftl.Topic[${2:Type}]("${1:topicName}")
