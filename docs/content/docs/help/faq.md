+++
title = "FAQ"
description = "Answers to frequently asked questions."
date = 2021-05-01T19:30:00+00:00
updated = 2021-05-01T19:30:00+00:00
draft = false
weight = 30
sort_by = "weight"
template = "docs/page.html"

[extra]
lead = "Answers to frequently asked questions."
toc = true
top = false
+++

## Why does FTL not allow external types?

Because of the nature of writing FTL verbs and data types, it's easy to think of it as just writing standard native code. Through that lens it is then somewhat surprising when FTL disallows the use of arbitrary external data types.

However, FTL types are not _just_ native types. FTL types are a more convenient method of writing an IDL such as [Protobufs](https://protobuf.dev/), [OpenAPI](https://www.openapis.org/) or [Thrift](https://thrift.apache.org/). With this in mind the constraint makes more sense. An IDL by its very nature must support a multitude of languages, so including an arbitrary type from a third party native library in one language may not be translatable to another language.

There are also secondary reasons, such as:

- Unclear ownership - in FTL a type must be owned by a single module. When importing a common third party type from multiple modules, which module owns that type?
- An external type must be representable in the FTL schema. The schema is then used to generate types for other modules, including those in other languages. For an external type, FTL could track the external library it belongs to to generate the "correct" code, but this would only be representable in a single language.
- External types often perform custom marshalling to/from JSON. This is not representable cross-language.
- Cleaner separation of abstraction layers - the ability to mix in abitrary external types is convenient, but can easily lead to mixing of concerns between internal and external data representations.

So what to do? While there are good reasons to disallow external types, it's also very irritating to have to manually transcribe types, or translate between JSON "blobs" in FTL and strong internal types. We're not sure how, but we definitely want to improve this experience. There is a draft [design document](/S8iS08PFT4SdnXzs4BIt8A) enumerating some options, please add your thoughts.

## What is a "module"?

In its least abstract form, a module is a collection of verbs, and the resources (databases, queues, cron jobs, secrets, config, etc.) that those verbs rely on to operate. All resources are private to their owning module.

More abstractly, the separation of concerns between modules is largely subjective. You _can_ think of each module as largely analogous to a traditional service, so when asking where the division between modules is that could inform your decision. That said, the ease of deploying modules in FTL is designed to give you more flexibility in how you structure your code.

## How do I represent optional/nullable values?

FTL's type system includes support for optionals. In Go this is represented as `ftl.Option[T]`, in languages with first-class support for optionals such as Kotlin, FTL will leverage the native type system.

When FTL is mapping to JSON, optional values will be represented as `null`.

In Go specifically, pointers to values are not supported because pointers are semantically ambiguous and error prone. They can mean, variously: "this value may or may not be present", or "this value just happens to be a pointer", or "this value is a pointer because it's mutable"

Additionally pointers to builtin types are painful to use in Go because you can't obtain a reference to a literal.

## Why must requests/responses be data structures, can't they be arrays, etc.?

This is currently due to FTL relying on traditional [schema evolution](https://softwaremill.com/schema-evolution-protobuf-scalapb-fs2grpc/) for forwards/backwards compatibility - eg. changing a slice to a struct in a backward compatible way is not possible, as an existing deployed peer consuming the slice will fail if it suddenly changes to a data structure.

Eventually FTL will allow multiple versions of a verb to be simultaneously deployed, such that a version returning a slice can coexist temporarily with a version returning a struct. Once all peers have been updated to support the new type signature, the old version will be dropped.

## I can't export a Verb from a nested package inside a subdirectory of the module root. What do I do?

Verbs and types can only be exported from the top level of each module. You are welcome to put any helper code you'd like in a nested package/directory.

## What types are supported by FTL?

FTL supports the following types: `Int` (64-bit), `Float` (64-bit), `String`, `Bytes` (a byte array), `Bool`, `Time`, `Any` (a dynamic type), `Unit` (similar to "void"), arrays, maps, data structures, and constant enumerations. Each FTL type is mapped to a corresponding language-specific type. For example in Go `Float` is represented as `float64`, `Time` is represented by `time.Time`, and so on.

Note that currently (until [type widening](https://github.com/TBD54566975/ftl/issues/1296) is implemented), external types are not supported.

## SQL errors on startup?

For example:

```bash
# ftl dev                                                                                            ~/src/ftl
info: Starting FTL with 1 controller(s)
ftl: error: ERROR: relation "fsm_executions" does not exist (SQLSTATE 42P01)
```

Run again with `ftl dev --recreate`. This usually indicates that your DB has an old schema.

This can occur when FTL has been upgraded with schema changes, making the database out of date. While in alpha we do not use schema migrations, so this won't occur once we hit a stable release.
