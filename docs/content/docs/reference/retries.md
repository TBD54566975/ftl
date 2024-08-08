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

Any verb called asynchronously (specifically, PubSub subscribers and FSM states), may optionally specify a basic exponential backoff retry policy via a Go comment directive. The directive has the following syntax:

```go
//ftl:retry [<attempts=10>] <min-backoff> [<max-backoff=1hr>] [catch <catchVerb>]
```

For example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.

```go
//ftl:retry 10 5s 1m
func Invoiced(ctx context.Context, in Invoice) error {
  // ...
}
```

## Catching
After all retries have failed, a catch verb can be used to safely recover.

These catch verbs have a request type of `builtin.CatchRequest<Req>` and no response type. If a catch verb returns an error, it will be retried until it succeeds so it is important to handle errors carefully.

```go
//ftl retry 5 1s catch recoverPaymentProcessing
func ProcessPayment(ctx context.Context, payment Payment) error {
    ...
}

//ftl:verb
func RecoverPaymentProcessing(ctx context.Context, request builtin.CatchRequest[Payment]) error {
    // safely handle final failure of the payment
}
```

For FSMs, after a catch verb has been successfully called the FSM will moved to the failed state.