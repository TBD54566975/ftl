+++
title = "PubSub"
description = "Asynchronous publishing of events to topics"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 80
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

FTL has first-class support for PubSub, modelled on the concepts of topics (where events are sent), subscriptions (a cursor over the topic), and subscribers (functions events are delivered to). Subscribers are, as you would expect, sinks. Each subscription is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each subscription may have multiple subscribers, in which case events will be distributed among them.

First, declare a new topic:

```go
var invoicesTopic = ftl.Topic[Invoice]("invoices")
```

Then declare each subscription on the topic:

```go
var _ = ftl.Subscription(invoicesTopic, "emailInvoices")
```

And finally define a Sink to consume from the subscription:

```go
//ftl:subscribe emailInvoices
func SendInvoiceEmail(ctx context.Context, in Invoice) error {
  // ...
}
```

Events can be published to a topic like so:

```go
invoicesTopic.Publish(ctx, Invoice{...})
```

> **NOTE!**
> PubSub topics cannot be published to from outside the module that declared them, they can only be subscribed to. That is, if a topic is declared in module `A`, module `B` cannot publish to it.
