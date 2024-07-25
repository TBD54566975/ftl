+++
title = "Glossary"
description = "Glossary of terms and definitions in FTL"
date = 2021-05-01T19:30:00+00:00
updated = 2021-05-01T19:30:00+00:00
draft = false
weight = 40
sort_by = "weight"
template = "docs/page.html"

[extra]
lead = "Glossary of terms and definitions in FTL."
toc = true
top = false
+++

##### Verb

A Verb is a remotely callable function that takes an input and returns an output.

```go
func(context.Context, In) (Out, error)
```

##### Sink

A Sink is a function that takes an input and returns nothing.

```go
func(context.Context, In) error
```

##### Source

A Source is a function that takes no input and returns an output.

```go
func(context.Context) (Out, error)
```

##### Empty

An Empty function is one that takes neither input or output.

```go
func(context.Context) error
```
