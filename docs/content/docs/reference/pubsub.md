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

{% code_selector() %}
<!-- go -->

First, declare a new topic:

```go
var Invoices = ftl.Topic[Invoice]("invoices")
```

Then declare each subscription on the topic:

```go
var _ = ftl.Subscription(Invoices, "emailInvoices")
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
Invoices.Publish(ctx, Invoice{...})
```

<!-- kotlin -->

First, declare a new topic :

```kotlin
@Export
@TopicDefinition("invoices")
internal interface InvoiceTopic : Topic<Invoice>
```

Events can be published to a topic by injecting it into an `@Verb` method:

```kotlin
@Verb
fun publishInvoice(request: InvoiceRequest, topic: InvoiceTopic) {
    topic.publish(Invoice(request.getInvoiceNo()))
}
```

There are two ways to subscribe to a topic. The first is to declare a method with the `@Subscription` annotation, this is generally used when
subscribing to a topic inside the same module:

```kotlin
@Subscription(topic = "invoices", name = "invoicesSubscription")
fun consumeInvoice(event: Invoice) {
    // ...
}
```

This is ok, but it requires the use of string constants for the topic name, which can be error-prone. If you are subscribing to a topic from
another module, FTL will generate a type-safe subscription meta annotation you can use to subscribe to the topic:

```kotlin
@Subscription(topic = "invoices", module = "publisher", name = "invoicesSubscription")
annotation class InvoicesSubscription 
```

This annotation can then be used to subscribe to the topic:

```kotlin
@InvoicesSubscription
fun consumeInvoice(event: Invoice) {
    // ...
}
```

Note that if you want multiple subscriptions or control over the subscription name you will need to use the `@Subscription` annotation.

<!-- java -->

First, declare a new topic:

```java
@Export
@TopicDefinition("invoices")
interface InvoiceTopic extends Topic<Invoice> {}
```

Events can be published to a topic by injecting it into an `@Verb` method:

```java
@Verb
void publishInvoice(InvoiceRequest request, InvoiceTopic topic) throws Exception {
    topic.publish(new Invoice(request.getInvoiceNo()));
}
```

There are two ways to subscribe to a topic. The first is to declare a method with the `@Subscription` annotation, this is generally used when
subscribing to a topic inside the same module:

```java
@Subscription(topic = "invoices", name = "invoicesSubscription")
public void consumeInvoice(Invoice event) {
    // ...
}
```

This is ok, but it requires the use of string constants for the topic name, which can be error-prone. If you are subscribing to a topic from
another module, FTL will generate a type-safe subscription meta annotation you can use to subscribe to the topic:

```java
@Retention(java.lang.annotation.RetentionPolicy.RUNTIME)
@Subscription(
        topic = "invoices",
        module = "publisher",
        name = "invoicesSubscription"
)
public @interface InvoicesSubscription {
}
```

This annotation can then be used to subscribe to the topic:

```java
@InvoicesSubscription
public void consumeInvoice(Invoice event) {
    // ...
}
```

Note that if you want multiple subscriptions or control over the subscription name you will need to use the `@Subscription` annotation.

{% end %}
> **NOTE!**
> PubSub topics cannot be published to from outside the module that declared them, they can only be subscribed to. That is, if a topic is declared in module `A`, module `B` cannot publish to it.
