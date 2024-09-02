+++
title = "FSM"
description = "Distributed Finite-State Machines"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 90
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

FTL has first-class support for distributed [finite-state machines](https://en.wikipedia.org/wiki/Finite-state_machine). Each state in the state machine is a Sink, with events being values of the type of each sinks input. The FSM is declared once, with each executing instance of the FSM identified by a unique key when sending an event to it.

Here's an example of an FSM that models a simple payment flow:

```go
var payment = ftl.FSM(
  "payment",
  ftl.Start(Invoiced),
  ftl.Start(Paid),
  ftl.Transition(Invoiced, Paid),
  ftl.Transition(Invoiced, Defaulted),
)

//ftl:verb
func SendDefaulted(ctx context.Context, in DefaultedInvoice) error {
  return payment.Send(ctx, in.InvoiceID, in.Timeout)
}

//ftl:verb
func Invoiced(ctx context.Context, in Invoice) error {
  if timedOut {
    return ftl.CallAsync(ctx, SendDefaulted, Timeout{...})
  }
}

//ftl:verb
func Paid(ctx context.Context, in Receipt) error { /* ... */ }

//ftl:verb
func Defaulted(ctx context.Context, in Timeout) error { /* ... */ }
```

## Creating and transitioning instances

To send an event to an fsm instance, call `Send()` on the FSM with the instance's unique key. The first time you send an event for an instance key, an fsm instance will be created. 

An example of creating an FSM instance and then transitioning it through it's states is shown below:

```go
err := payment.Send(ctx, invoiceID, Invoice {Amount: 110})
err = payment.Send(ctx, invoiceID, Receipt {Amount: 110})
```

When an event is sent to the FSM the method to be called is determined by matching the current state and event payload
type to methods that can transition from the current state and have the same payload type. In the example above the first
`Send` call will created the FSM, and will call the `Invoiced` method as it is a start state and takes an `Invoice` as
payload. The second `Send` call will call the `Paid` method as it is a transition from the `Invoiced` state and takes a
`Receipt` as payload. If the second call had sent a `Timeout` instead of a `Receipt` the FSM would have called the `Defaulted`
method instead.

It is important to note that in this model the methods both represent a state, and a way to transition into that
state. This means when a method is invoked it always moves to the corresponding state, consider the following example:

```go
err := payment.Send(ctx, invoiceID, Invoice {Amount: 110})
err = payment.Send(ctx, invoiceID, Receipt {Amount: 20})
```

In this case it would still moved to the `Paid` state even though the customer only paid 20 of the 110.

Sending an event to an FSM is asynchronous. From the time an event is sent until the state function completes execution, the FSM is transitioning. It is invalid to send an event to an FSM that is transitioning.

During a transition you may need to trigger a transition to another state. This can be done by calling `Next()` on the FSM:

```go
err := payment.Next(ctx, invoiceID, Receipt{...})
```