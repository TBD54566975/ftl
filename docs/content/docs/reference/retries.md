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
//ftl:retry [<attempts>] <min-backoff> [<max-backoff>]
```

`attempts` and `max-backoff` default to unlimited if not specified.

For example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.

```go
//ftl:retry 10 5s 1m
func Invoiced(ctx context.Context, in Invoice) error {
  // ...
}
```
