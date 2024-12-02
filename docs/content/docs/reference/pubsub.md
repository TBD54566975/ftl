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

FTL has first-class support for PubSub, modelled on the concepts of topics (where events are sent) and subscribers (a verb which consumes events). Subscribers are, as you would expect, sinks. Each subscriber is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each published event has an at least once delivery guarantee for each subscription.

{% code_selector() %}
<!-- go -->

First, declare a new topic:

```go
package payments

var Invoices = ftl.Topic[Invoice]("invoices")
```

Then define a Sink to consume from the topic:

```go
//ftl:subscribe payments.invoices from=beginning
func SendInvoiceEmail(ctx context.Context, in Invoice) error {
  // ...
}
```

Events can be published to a topic like so:

```go
Invoices.Publish(ctx, Invoice{...})
```

<!-- kotlin -->

First, declare a new topic :

```kotlin
@Export
@Topic("invoices")
internal interface InvoiceTopic : WriteableTopic<Invoice>
```

Events can be published to a topic by injecting it into an `@Verb` method:

```kotlin
@Verb
fun publishInvoice(request: InvoiceRequest, topic: InvoiceTopic) {
    topic.publish(Invoice(request.getInvoiceNo()))
}
```

To subscribe to a topic use the `@Subscription` annotation, referencing the topic class and providing a method to consume the event:

```kotlin
@Subscription(topic = InvoiceTopic::class, from = FromOffset.LATEST)
fun consumeInvoice(event: Invoice) {
    // ...
}
```

If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated
topic cannot be published to, only subscribed to:

```kotlin
@Topic(name="invoices", module="publisher")
internal interface InvoiceTopic : ConsumableTopic<Invoice>
```

<!-- java -->

First, declare a new topic:

```java
@Export
@Topic("invoices")
interface InvoiceTopic extends WriteableTopic<Invoice> {}
```

Events can be published to a topic by injecting it into an `@Verb` method:

```java
@Verb
void publishInvoice(InvoiceRequest request, InvoiceTopic topic) throws Exception {
    topic.publish(new Invoice(request.getInvoiceNo()));
}
```

To subscribe to a topic use the `@Subscription` annotation, referencing the topic class and providing a method to consume the event:

```java
@Subscription(topic = InvoiceTopic.class, from = FromOffset.LATEST)
public void consumeInvoice(Invoice event) {
    // ...
}
```

If you are subscribing to a topic from another module, FTL will generate a topic class for you so you can subscribe to it. This generated
topic cannot be published to, only subscribed to:

```java
@Topic(name="invoices", module="publisher")
 interface InvoiceTopic extends ConsumableTopic<Invoice> {}
```

{% end %}
> **NOTE!**
> PubSub topics cannot be published to from outside the module that declared them, they can only be subscribed to. That is, if a topic is declared in module `A`, module `B` cannot publish to it.
