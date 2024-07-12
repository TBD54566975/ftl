+++
title = "External Types"
description = "Using external types in your modules"
date = 2024-07-12T18:00:00+00:00
updated = 2024-07-12T18:00:00+00:00
draft = false
weight = 110
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

## Using external types

To use an external type in your FTL module schema, declare a type alias over the external type:

```go
//ftl:typealias
type FtlType external.OtherType
```

The external type is widened to `Any` in the FTL schema, and the corresponding type alias will include metadata 
for the runtime-specific type mapping:

```
typealias FtlType Any
  +typemap go "github.com/external.OtherType"
```

Users can achieve functionally equivalent behavior to using the external type directly by using the declared 
alias (`FtlType`) throughout their code. Direct usage of the external type in schema declarations is not supported; 
instead, the type alias must be used.

FTL will automatically serialize and deserialize the external type to the strong type indicated by the mapping.

## Cross-Runtime Type Mappings

FTL also provides the capability to declare type mappings for other runtimes. For instance, to include a type mapping for Kotlin, you can 
annotate your type alias declaration as follows:

```go
//ftl:typealias
//ftl:typemap kotlin "com.external.other.OtherType"
type FtlType external.OtherType
```

In the FTL schema, this will appear as:

```
typealias FtlType Any
  +typemap go "github.com/external.OtherType"
  +typemap kotlin "com.external.other.OtherType"
```

This allows FTL to decode the type properly in other languages, for seamless 
interoperability across different runtimes.