+++
title = "Retries"
description = "Retrying asynchronous verbs"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 100
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

Some FTL features allow specifying a retry policy via a Go comment directive. Retries back off exponentially until the maximum is reached.

The directive has the following syntax:

```go
//ftl:retry [<attempts=10>] <min-backoff> [<max-backoff=1hr>] [catch <catchVerb>]
```

For example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.

```go
//ftl:retry 10 5s 1m
func Process(ctx context.Context, in Invoice) error {
  // ...
}
```

### PubSub

Subscribers can have a retry policy, like this:
```go
//ftl:subscribe exampleSubscription
//ftl:retry 5 1s catch recoverPaymentProcessing
func ProcessPayment(ctx context.Context, payment Payment) error {
    ...
}
```

### FSM

Retries can be declared on the FSM or on individual transition verbs. Retries declared on a verb take precedence over ones declared on the FSM. For example:
```go
//ftl:retry 10 1s 10s
var fsm = ftl.FSM("fsm",
	ftl.Start(Start),
	ftl.Transition(Start, End),
)

//ftl:verb
//ftl:retry 1 1s 1s
func Start(ctx context.Context, in Event) error {
	// Start uses its own retry policy
}


//ftl:verb
func End(ctx context.Context, in Event) error {
	// End inherits the default retry policy from the FSM
}
```


## Catching
After all retries have failed, a catch verb can be used to safely recover.

These catch verbs have a request type of `builtin.CatchRequest<Req>` and no response type. If a catch verb returns an error, it will be retried until it succeeds so it is important to handle errors carefully.

```go
//ftl:retry 5 1s catch recoverPaymentProcessing
func ProcessPayment(ctx context.Context, payment Payment) error {
    ...
}

//ftl:verb
func RecoverPaymentProcessing(ctx context.Context, request builtin.CatchRequest[Payment]) error {
    // safely handle final failure of the payment
}
```

For FSMs, after a catch verb has been successfully called the FSM will moved to the failed state.