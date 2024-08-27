+++
title = "Start"
description = "Preparing to use FTL."
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 10
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

## Import the runtime

{% code_selector() %}

Some aspects of FTL rely on a runtime which must be imported with:

```go
import "github.com/TBD54566975/ftl/go-runtime/ftl"
```

```kotlin
// Declare a verb you want to use:
import xyz.block.ftl.Verb

// When using the export feature:
import xyz.block.ftl.Export
```

{% end %}
